// Package csfle implements a boilerplate code for the creation of a CSFLE mongoDB client
package csfle

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	m "github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"
)

func New(ctx context.Context, serviceName string, logger log.Log, c Config, t m.Tracer, opts ...*options.ClientOptions) (*m.Mongo, error) {
	extraOptions := map[string]interface{}{
		"cryptSharedLibPath":     c.CryptSharedLibPath,
		"mongocryptdBypassSpawn": true,
		"cryptSharedLibRequired": true,
	}
	autoEncryptionOpts := options.AutoEncryption().
		SetKmsProviders(c.KMSCredentials).
		SetKeyVaultNamespace(c.KeyVaultNamespace).
		SetEncryptedFieldsMap(c.SchemaMap).
		SetExtraOptions(extraOptions)
	connectionOptions := options.Client()
	connectionOptions.SetAutoEncryptionOptions(autoEncryptionOpts)
	opts = utils.Prepend(opts, connectionOptions)
	return m.NewWithDefaultOptions(ctx, serviceName, logger, *c.Config, t, opts...)
}

func GetDataKey(ctx context.Context, m *m.Mongo, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (res *primitive.Binary, err error) {

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
			err = fmt.Errorf("csfle.GetDataKey: error finding data key: %w", err)
			return
		}
		res, err = CreateDataKey(ctx, m, keyVaultNamespace, keyAltName, provider)
		if err != nil {
			return
		}
	} else {
		data := dataKey["_id"].(primitive.Binary)
		res = &data
	}
	return

}

func CreateDataKey(ctx context.Context, m *m.Mongo, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (*primitive.Binary, error) {
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
	return &dataKeyID, nil
}
