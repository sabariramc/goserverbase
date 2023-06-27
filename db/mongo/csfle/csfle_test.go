package csfle_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sabariramc/goserverbase/v3/db/mongo"
	"github.com/sabariramc/goserverbase/v3/db/mongo/csfle"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		t.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	schemeByte, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	scheme := string(schemeByte)
	kmsArn := MongoTestConfig.AWS.KMS_ARN
	keyAltName := "MongoPIITestKey"
	kmsProvider, err := csfle.GetDefaultAWSKMSProvider(ctx, MongoTestLogger, kmsArn)
	if err != nil {
		t.Fatal(err)
	}
	keyNamespace := "__TestNameSpace.__Coll"
	err = csfle.SetEncryptionKey(ctx, MongoTestLogger, &scheme, *MongoTestConfig.Mongo, keyNamespace, keyAltName, kmsProvider)
	if err != nil {
		t.Fatal(err)
	}
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		t.Fatal(err)
	}
	coll := client.Database("GOTEST").Collection("PII")
	mongoScheme, err := csfle.CreateBSONSchema(&scheme, "GOLANGTEST", "PII")
	if err != nil {
		t.Fatal(err)
	}
	csfleMongoClient, err := csfle.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo, keyNamespace, mongoScheme, kmsProvider)
	csfleClient := mongo.NewWrapper(ctx, MongoTestLogger, *MongoTestConfig.Mongo, csfleMongoClient)
	piicoll := csfleClient.Database("GOTEST").Collection("PII")
	piicoll.SetHashList([]string{"pan", "email"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = piicoll.InsertOneWithHash(ctx, map[string]interface{}{
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
		t.Fatal(err)
	}
	cur := piicoll.FindOneWithHash(ctx, map[string]interface{}{"email": "sab@sabariram.com"})
	val := &PIITestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	piicoll.UpdateByIDWithHash(ctx, val.ID, map[string]map[string]interface{}{"$set": {"email": "iam2@gosabariram.com"}})
	cur = piicoll.FindOneWithHash(ctx, map[string]interface{}{"email": "iam2@gosabariram.com"})
	val = &PIITestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	cur = piicoll.FindOne(ctx, map[string]interface{}{"_id": val.ID})
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}

	data, err := coll.Find(ctx, map[string]map[string]interface{}{"pan": {"$exists": true}})
	if err != nil {
		t.Fatal(err)
	}
	for data.Next(ctx) {
		decodeData := make(map[string]interface{})
		data.Decode(&decodeData)
		fmt.Printf("%+v\n", decodeData)
	}
	res, err := piicoll.DeleteOne(ctx, map[string]interface{}{"_id": val.ID})
	if err != nil {
		t.Fatal(err)
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

	_, err = piicoll.InsertOneWithHash(ctx, map[string]interface{}{
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
		t.Fatal(err)
	}

}
