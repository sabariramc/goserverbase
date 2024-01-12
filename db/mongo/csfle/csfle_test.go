package csfle_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sabariramc/goserverbase/v4/db/mongo"
	"github.com/sabariramc/goserverbase/v4/db/mongo/csfle"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gotest.tools/assert"
)

type Address struct {
	AddressLine1 string `bson:"addressLine1"`
	AddressLine2 string `bson:"addressLine2"`
	AddressLine3 string `bson:"addressLine3"`
	State        string `bson:"state"`
	PIN          string `bson:"pin"`
	Country      string `bson:"country"`
}

type PIITestVal struct {
	ID      primitive.ObjectID `bson:"_id"`
	DOB     string             `bson:"dob"`
	Name    string             `bson:"name"`
	Pan     string             `bson:"pan"`
	Email   string             `bson:"email"`
	Address Address            `bson:"address"`
}

func TestCollectionPII(t *testing.T) {
	ctx := GetCorrelationContext()
	file, err := os.Open("./sample/piischeme.json")
	if err != nil {
		assert.NilError(t, err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			assert.NilError(t, err)
		}
	}()
	schemeByte, err := io.ReadAll(file)
	if err != nil {
		assert.NilError(t, err)
	}
	scheme := string(schemeByte)
	kmsArn := MongoTestConfig.AWS.KMS_ARN
	keyAltName := "MongoPIITestKey"
	kmsProvider, err := csfle.GetDefaultAWSKMSProvider(ctx, MongoTestLogger, kmsArn)
	if err != nil {
		assert.NilError(t, err)
	}
	keyNamespace := "__TestNameSpace.__Coll"
	err = csfle.SetEncryptionKey(ctx, MongoTestLogger, &scheme, *MongoTestConfig.Mongo, keyNamespace, keyAltName, kmsProvider)
	if err != nil {
		assert.NilError(t, err)
	}
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		assert.NilError(t, err)
	}
	coll := client.Database("GOTEST").Collection("PII")
	mongoScheme, err := csfle.CreateBSONSchema(&scheme, "GOTEST", "PII")
	if err != nil {
		assert.NilError(t, err)
	}
	csfleMongoClient, err := csfle.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo, keyNamespace, mongoScheme, kmsProvider)
	csfleClient := mongo.NewWrapper(ctx, MongoTestLogger, csfleMongoClient)
	piicoll := csfleClient.Database("GOTEST").Collection("PII")
	if err != nil {
		assert.NilError(t, err)
	}
	uuid := "FAsdfasfsadfsdafs"
	_, err = piicoll.InsertOne(ctx, map[string]interface{}{
		"dob":   "1991-08-02",
		"name":  "Sabariram",
		"pan":   "ABCDE1234F",
		"email": "sab@sabariram.com",
		"address": map[string]string{
			"addressLine1": "door no with street name",
			"addressLine2": "taluk and postal office",
			"addressLine3": "Optional landmark",
			"state":        "Tamil Nadu",
			"pin":          "636351",
			"country":      "India",
		},
		"UUID": uuid,
	})
	if err != nil {
		assert.NilError(t, err)
	}
	cur := piicoll.FindOne(ctx, map[string]interface{}{"UUID": uuid})
	val := &PIITestVal{}
	err = cur.Decode(val)
	if err != nil {
		assert.NilError(t, err)
	}
	fmt.Printf("%+v\n", val)
	piicoll.UpdateOne(ctx, val.ID, map[string]map[string]interface{}{"$set": {"UUID": uuid}})
	cur = piicoll.FindOne(ctx, map[string]interface{}{"UUID": uuid})
	val = &PIITestVal{}
	err = cur.Decode(val)
	if err != nil {
		assert.NilError(t, err)
	}
	fmt.Printf("%+v\n", val)
	cur = piicoll.FindOne(ctx, map[string]interface{}{"_id": val.ID})
	err = cur.Decode(val)
	if err != nil {
		assert.NilError(t, err)
	}

	data, err := coll.Find(ctx, map[string]map[string]interface{}{"pan": {"$exists": true}})
	if err != nil {
		assert.NilError(t, err)
	}
	for data.Next(ctx) {
		decodeData := make(map[string]interface{})
		data.Decode(&decodeData)
		fmt.Printf("%+v\n", decodeData)
	}
	res, err := piicoll.DeleteOne(ctx, map[string]interface{}{"_id": val.ID})
	if err != nil {
		assert.NilError(t, err)
	}
	if res.DeletedCount != 1 {
		t.Fatal("Delete count is not matching")
	}
	cur = piicoll.FindOne(ctx, map[string]interface{}{"_id": val.ID})
	err = cur.Decode(val)
	if err != nil {
		fmt.Println(err)
	} else {
		t.Fatal(fmt.Errorf("doc shouldn't exist"))
	}
	_, err = piicoll.InsertOne(ctx, map[string]interface{}{
		"dob":   "1991-08-02",
		"name":  "Vamshi Krishna",
		"pan":   "ABCDE1234F",
		"email": "sab@sabariram.com",
		"address": map[string]string{
			"addressLine1": "door no with street name",
			"addressLine2": "taluk and postal office",
			"addressLine3": "Optional landmark",
			"state":        "Tamil Nadu",
			"pin":          "636351",
			"country":      "India",
		},
		"UUID": "FAsdfasfsadfsdafs",
	})
	if err != nil {
		assert.NilError(t, err)
	}

}
