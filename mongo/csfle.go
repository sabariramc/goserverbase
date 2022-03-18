package mongo

import (
	"context"
	"encoding/json"
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

func NewDefaultCSFLECollection(ctx context.Context, mongoURI, keyVaultNamespace, databaseName, collectionName, encryptionSchema string, kmsProvider MasterKeyProvider, hashFieldList []string) (*Collection, error) {
	logger := log.GetDefaultLogger()
	return NewCSFLECollection(ctx, mongoURI, keyVaultNamespace, databaseName, collectionName, encryptionSchema, kmsProvider, logger, hashFieldList)
}

func NewCSFLECollection(ctx context.Context, mongoURI, keyVaultNamespace, databaseName, collectionName, encryptionSchema string, kmsProvider MasterKeyProvider, logger *log.Log, hashFieldList []string) (*Collection, error) {
	clientKey := fmt.Sprintf("%v_%v_%v", mongoURI, databaseName, collectionName)
	uriHash := utils.GetHash(clientKey)
	cachedClient, ok := piiCache[uriHash]
	var client *mongo.Client
	if !ok || time.Now().After(cachedClient.expireTime) {
		schema, err := CreateBSONSchema(&encryptionSchema, databaseName, collectionName)
		if err != nil {
			logger.Error("Error in creating CLFLE scheme", err)
			return nil, err
		}
		client, err = NewCSFLEClient(ctx, keyVaultNamespace, mongoURI, schema, kmsProvider, logger)
		if err != nil {
			return nil, err
		}
		piiCache.cache(uriHash, client)
	} else {
		logger.Debug("Mongo CSFLE client taken from cache", nil)
		client = cachedClient.client
	}
	coll := NewCollection(ctx, client, databaseName, collectionName, logger)
	coll.SetHashList(hashFieldList)
	return coll, nil
}

func SetEncryptionKey(ctx context.Context, encryptionSchema *string, mongoURI, keyVaultNamespace, keyAltName string, logger *log.Log, kmsProvider MasterKeyProvider) error {
	schema := make(map[string]interface{})
	err := json.Unmarshal([]byte(*encryptionSchema), &schema)
	if err != nil {
		logger.Error("CSFL Schema unmarshal error", err)
		return err
	}
	client, err := GetDefaultClient(ctx, mongoURI)
	if err != nil {
		return err
	}
	keyID, err := client.GetDataKey(keyVaultNamespace, keyAltName, kmsProvider)
	if err != nil {
		return err
	}
	encryptMetadataIn, ok := schema["encryptMetadata"]
	encryptionKeyID := []interface{}{
		map[string]interface{}{"$binary": map[string]interface{}{"base64": keyID, "subType": "04"}},
	}
	if ok {
		encryptMetadata, ok := encryptMetadataIn.(map[string]interface{})
		if !ok {
			errorMsg := "key `encryptMetadata` should be a compatible with `map[string]interface{}` in param `encryptionSchema`"
			logger.Error(errorMsg, schema)
			return fmt.Errorf(errorMsg)
		}
		encryptMetadata["keyId"] = encryptionKeyID
	} else {
		schema["encryptMetadata"] = map[string]interface{}{
			"keyId": encryptionKeyID,
		}
	}
	blob, err := json.Marshal(schema)
	if err != nil {
		logger.Error("CSFL Schema marshal error", err)
		return err
	}
	*encryptionSchema = string(blob)
	return nil
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

func NewCSFLEClient(ctx context.Context, keyVaultNamespace, mongoURI string, schemaMap map[string]interface{}, provider MasterKeyProvider, logger *log.Log) (*mongo.Client, error) {
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
		logger.Error("connect error for Mongo CSLFE client", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Error pinging mongo server", err)
		return nil, err
	}
	return client, nil
}

func init() {
	err := os.MkdirAll(tmp, 0700)
	fmt.Printf("Folder Creation : %v\n", err)
}
