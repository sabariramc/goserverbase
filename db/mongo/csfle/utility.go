package csfle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/log"
)

func SetEncryptionKey(ctx context.Context, logger log.Log, encryptionSchema *string, client *mongo.Mongo, keyVaultNamespace string, kmsProvider MasterKeyProvider) (map[string]interface{}, error) {
	schema := make(map[string]interface{})
	err := json.Unmarshal([]byte(*encryptionSchema), &schema)
	if err != nil {
		return nil, fmt.Errorf("csfle.SetEncryptionKey: error unmarshalling schema: %w", err)
	}

	if err != nil {
		return nil, err
	}

	for coll, iConfig := range schema {
		collConfig, ok := iConfig.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("csfle.SetEncryptionKey: error parsing schema: invalid structure for collection - %v", coll)
		}
		fields, ok := collConfig["fields"].([]any)
		if !ok {
			return nil, fmt.Errorf("csfle.SetEncryptionKey: error parsing schema: invalid structure for fields in collection %v", coll)
		}
		for i := 0; i < len(fields); i++ {
			data, ok := fields[i].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("csfle.SetEncryptionKey: error parsing schema: invalid structure for fields in collection %v", coll)
			}
			keyAltName, ok := data["keyAltName"].(string)
			if !ok {
				return nil, fmt.Errorf("csfle.SetEncryptionKey: error parsing schema: keyAltName missing for field in collection %v", coll)
			}
			keyID, err := GetDataKey(ctx, client, keyVaultNamespace, keyAltName, kmsProvider)
			if err != nil {
				return nil, fmt.Errorf("csfle.SetEncryptionKey: keyFetchingError: %w", err)
			}
			delete(data, "keyAltName")
			data["keyId"] = keyID
		}
	}
	return schema, nil
}
