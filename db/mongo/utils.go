package mongo

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func CreateBSONSchema(schema *string, databaseName, collectionName string) (map[string]interface{}, error) {
	var schemaDoc bson.Raw
	if err := bson.UnmarshalExtJSON([]byte(*schema), true, &schemaDoc); err != nil {
		return nil, fmt.Errorf("mongo.CreateBSONSchema : %w", err)
	}
	return map[string]interface{}{
		databaseName + "." + collectionName: schemaDoc,
	}, nil
}
