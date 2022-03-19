package testutils

import (
	"context"
	"encoding/json"
	"fmt"

	"sabariram.com/goserverbase/db/mongo"
	"sabariram.com/goserverbase/log"
)

func SetEncryptionKey(ctx context.Context, logger *log.Logger, encryptionSchema *string, mongoURI, keyVaultNamespace, keyAltName string, kmsProvider mongo.MasterKeyProvider) error {
	schema := make(map[string]interface{})
	err := json.Unmarshal([]byte(*encryptionSchema), &schema)
	if err != nil {
		logger.Error(ctx, "CSFL Schema unmarshal error", err)
		return err
	}
	client, err := mongo.GetDefaultClient(ctx, logger, mongoURI)
	if err != nil {
		return err
	}
	keyID, err := client.GetDataKey(ctx, keyVaultNamespace, keyAltName, kmsProvider)
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
			logger.Error(ctx, errorMsg, schema)
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
		logger.Error(ctx, "CSFL Schema marshal error", err)
		return err
	}
	*encryptionSchema = string(blob)
	return nil
}
