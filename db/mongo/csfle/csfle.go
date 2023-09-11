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

	m "github.com/sabariramc/goserverbase/v3/db/mongo"
	"github.com/sabariramc/goserverbase/v3/log"
)

var tmp = "/tmp/mongocryptd"

func New(ctx context.Context, logger *log.Logger, c m.Config, keyVaultNamespace string, schemaMap map[string]interface{}, provider MasterKeyProvider, opts ...*options.ClientOptions) (*mongo.Client, error) {
	extraOptions := map[string]interface{}{
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
	opts = append(opts, connectionOptions)
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("csfle.NewCSFLEClient: error creating Mongo CSLFE client: %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("csfle.NewCSFLEClient: error pinging mongo server: %w", err)
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
			err = fmt.Errorf("csfle.GetDataKey: error encountered while attempting to find data key: %w", err)
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
		return nil, fmt.Errorf("csfle.CreateDataKey: error creating encryption client: %w", err)
	}
	dataKeyOpts := options.DataKey().
		SetMasterKey(provider.DataKeyOpts()).
		SetKeyAltNames([]string{keyAltName})
	dataKeyID, err := clientEnc.CreateDataKey(ctx, provider.Name(), dataKeyOpts)
	if err != nil {
		return nil, fmt.Errorf("csfle.CreateDataKey: error creating data key: %w", err)
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
