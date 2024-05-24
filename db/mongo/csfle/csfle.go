// Package csfle implements boilerplate code for creating a MongoDB client with Client-Side Field Level Encryption (CSFLE) capabilities.
package csfle

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	m "github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// New creates a new MongoDB client with CSFLE enabled using the provided configuration and client options.
func New(logger log.Log, t m.Tracer, c Config, opts ...*options.ClientOptions) (*m.Mongo, error) {
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
	return m.NewWithDefaultOptions(logger, t, opts...)
}

// GetDataKey retrieves the data encryption key from the key vault specified by keyVaultNamespace.
// If the data key does not exist, it creates a new one using the provided MasterKeyProvider.
func GetDataKey(ctx context.Context, m *m.Mongo, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (res *primitive.Binary, err error) {
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

// CreateDataKey creates a new data encryption key in the key vault specified by keyVaultNamespace,
// using the provided MasterKeyProvider for the master key.
func CreateDataKey(ctx context.Context, m *m.Mongo, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (*primitive.Binary, error) {
	clientEncryptionOpts := options.ClientEncryption().
		SetKeyVaultNamespace(keyVaultNamespace).
		SetKmsProviders(provider.Credentials())
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
