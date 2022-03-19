package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sabariram.com/goserverbase/log"
	"sabariram.com/goserverbase/utils"
)

var tmp = "/tmp/mongocryptd"

var piiCache clientCache = make(map[string]*clientCacheData)

func NewDefaultCSFLECollection(ctx context.Context, logger *log.Logger, mongoURI, keyVaultNamespace, databaseName, collectionName, encryptionSchema string, kmsProvider MasterKeyProvider, hashFieldList []string) (*Collection, error) {
	return NewCSFLECollection(ctx, logger, mongoURI, keyVaultNamespace, databaseName, collectionName, encryptionSchema, kmsProvider, hashFieldList)
}

func NewCSFLECollection(ctx context.Context, logger *log.Logger, mongoURI, keyVaultNamespace, databaseName, collectionName, encryptionSchema string, kmsProvider MasterKeyProvider, hashFieldList []string) (*Collection, error) {
	clientKey := fmt.Sprintf("%v_%v_%v", mongoURI, databaseName, collectionName)
	uriHash := utils.GetHash(clientKey)
	cachedClient, ok := piiCache[uriHash]
	var client *mongo.Client
	if !ok || time.Now().After(cachedClient.expireTime) {
		schema, err := CreateBSONSchema(&encryptionSchema, databaseName, collectionName)
		if err != nil {
			logger.Error(ctx, "Error in creating CLFLE scheme", err)
			return nil, err
		}
		client, err = NewCSFLEClient(ctx, logger, keyVaultNamespace, mongoURI, schema, kmsProvider)
		if err != nil {
			return nil, err
		}
		piiCache.cache(uriHash, client)
	} else {
		logger.Debug(ctx, "Mongo CSFLE client taken from cache", nil)
		client = cachedClient.client
	}
	coll := NewCollection(logger, client, databaseName, collectionName)
	coll.SetHashList(hashFieldList)
	return coll, nil
}

func CreateBSONSchema(schema *string, databaseName, collectionName string) (map[string]interface{}, error) {
	var schemaDoc bson.Raw
	if err := bson.UnmarshalExtJSON([]byte(*schema), true, &schemaDoc); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		databaseName + "." + collectionName: schemaDoc,
	}, nil
}

func NewCSFLEClient(ctx context.Context, logger *log.Logger, keyVaultNamespace, mongoURI string, schemaMap map[string]interface{}, provider MasterKeyProvider) (*mongo.Client, error) {
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
	connectionOptions.ApplyURI(mongoURI)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 12)
	connectionOptions.SetAutoEncryptionOptions(autoEncryptionOpts)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI).SetAutoEncryptionOptions(autoEncryptionOpts))
	if err != nil {
		logger.Error(ctx, "connect error for Mongo CSLFE client", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(ctx, "Error pinging mongo server", err)
		return nil, err
	}
	return client, nil
}

func init() {
	err := os.MkdirAll(tmp, 0700)
	if err != nil {
		fmt.Printf("Error in folder creation (%v) : %v\n", tmp, err)
	}
}
