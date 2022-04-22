package csfle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabariramc/goserverbase/config"
	"github.com/sabariramc/goserverbase/db/mongo"
	"github.com/sabariramc/goserverbase/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/log"
)

func SetEncryptionKey(ctx context.Context, logger *log.Logger, encryptionSchema *string, c config.MongoConfig, csfleC config.MongoCFLEConfig, keyAltName string, kmsProvider mongo.MasterKeyProvider) error {
	schema := make(map[string]interface{})
	err := json.Unmarshal([]byte(*encryptionSchema), &schema)
	if err != nil {
		logger.Error(ctx, "CSFL Schema unmarshal error", err)
		return err
	}
	client, err := mongo.NewMongo(ctx, logger, c)
	if err != nil {
		return err
	}
	keyID, err := csfle.GetDataKey(ctx, client, csfleC.KeyVaultNamespace, keyAltName, kmsProvider)
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
