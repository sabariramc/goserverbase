package csfle

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"

	m "github.com/sabariramc/goserverbase/db/mongo"
	"github.com/sabariramc/goserverbase/log"
)

var tmp = "/tmp/mongocryptd"

func New(ctx context.Context, logger *log.Logger, c m.Config, keyVaultNamespace string, schemaMap map[string]interface{}, provider MasterKeyProvider) (*mongo.Client, error) {
	// extra options that can be specified
	// see https://github.com/mongodb/specifications/blob/master/source/client-side-encryption/client-side-encryption.rst#extraoptions

	extraOptions := map[string]interface{}{
		// "mongocryptdURI":, defaults to "mongodb://localhost:27020"
		// "mongocryptdBypassSpawn":, defaults to false
		// "mongocryptdSpawnPath": mongocryptPath + "/bin/mongocryptd",
		"mongocryptdSpawnArgs": []string{fmt.Sprintf("--pidfilepath=%v/mongocryptd.pid", tmp)},
	}
	autoEncryptionOpts := options.AutoEncryption().
		SetKmsProviders(provider.Credentials()).
		SetKeyVaultNamespace(keyVaultNamespace).
		SetSchemaMap(schemaMap).
		SetExtraOptions(extraOptions)
	connectionOptions := options.Client()
	connectionOptions.Monitor = mongotrace.NewMonitor()
	connectionOptions.ApplyURI(c.ConnectionString)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 12)
	connectionOptions.SetAutoEncryptionOptions(autoEncryptionOpts)
	connectionOptions.SetMinPoolSize(c.MinConnectionPool)
	connectionOptions.SetMaxPoolSize(c.MaxConnectionPool)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(c.ConnectionString).SetAutoEncryptionOptions(autoEncryptionOpts))
	if err != nil {
		logger.Error(ctx, "connect error for Mongo CSLFE client", err)
		return nil, fmt.Errorf("csfle.NewCSFLEClient : %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(ctx, "Error pinging mongo server", err)
		return nil, fmt.Errorf("csfle.NewCSFLEClient : %w", err)
	}
	return client, nil
}

func GetDataKey(ctx context.Context, m *m.Mongo, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (res string, err error) {

	// configuring encryption options by setting the keyVault namespace and the kms providers information
	// we configure this client to fetch the master key so that we can
	// create a data key in the next step

	// look for a data key
	keyVault := strings.Split(keyVaultNamespace, ".")
	db := keyVault[0]
	coll := keyVault[1]
	var dataKey bson.M
	err = m.GetClient().
		Database(db).
		Collection(coll).
		FindOne(ctx, bson.M{"keyAltNames": keyAltName}).
		Decode(&dataKey)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			m.GetLogger().Error(ctx, "error encountered while attempting to find key", err)
			return
		}
		var data *string
		data, err = CreateDataKey(ctx, m, keyVaultNamespace, keyAltName, provider)
		if err != nil {
			return
		}
		res = *data

	} else {
		res = base64.StdEncoding.EncodeToString(dataKey["_id"].(primitive.Binary).Data)
	}
	return

}

func CreateDataKey(ctx context.Context, m *m.Mongo, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (*string, error) {
	// specify the master key information that will be used to
	// encrypt the data key(s) that will in turn be used to encrypt
	// fields, and create the data key
	clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).SetKmsProviders(provider.Credentials())
	clientEnc, err := mongo.NewClientEncryption(m.GetClient(), clientEncryptionOpts)
	if err != nil {
		m.GetLogger().Error(ctx, "Error creating encryption client", err)
		return nil, fmt.Errorf("csfle.CreateDataKey : %w", err)
	}
	dataKeyOpts := options.DataKey().
		SetMasterKey(provider.DataKeyOpts()).
		SetKeyAltNames([]string{keyAltName})
	dataKeyID, err := clientEnc.CreateDataKey(ctx, provider.Name(), dataKeyOpts)
	if err != nil {
		m.GetLogger().Error(ctx, "Error creating data key", err)
		return nil, fmt.Errorf("csfle.CreateDataKey : %w", err)
	}
	res := base64.StdEncoding.EncodeToString(dataKeyID.Data)
	return &res, nil
}

func init() {
	err := os.MkdirAll(tmp, 0700)
	if err != nil {
		fmt.Printf("Error in folder creation (%v) : %v\n", tmp, err)
	}
}
