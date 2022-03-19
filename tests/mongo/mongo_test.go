package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sabariram.com/goserverbase/db/mongo"
	"sabariram.com/goserverbase/utils"
	"sabariram.com/goserverbase/utils/testutils"
)

var mongoURI = os.Getenv("MONGO_URL")

type TestVal struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	IntVal  int64              `bson:"intVal"`
	DeciVal decimal.Decimal    `bson:"deciVal"`
	StrVal  string             `bson:"strVal"`
	BoolVal bool               `bson:"boolVal"`
	TimeVal time.Time          `bson:"timeVal"`
}

func TestMongocollection(t *testing.T) {
	ctx := GetCorrelationContext()
	coll, err := mongo.NewDefaultCollection(ctx, MongoTestLogger, mongoURI, "GOLANGTEST", "Plain")
	if err != nil {
		t.Fatal(err)
	}
	val1, _ := decimal.NewFromString("123.1232")
	val2, _ := decimal.NewFromString("123.1232")
	if err != nil {
		t.Fatal(err)
	}
	coll.InsertOne(ctx, map[string]interface{}{
		"strVal":  "value1",
		"intVal":  123,
		"deciVal": val1.Add(val2),
		"timeVal": time.Now().In(utils.IST),
	})
	cur := coll.FindOne(ctx, map[string]string{"strVal": "value1"})
	val := &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	coll.UpdateByID(ctx, val.ID, map[string]map[string]interface{}{"$set": {"strVal": "val2"}})
	cur = coll.FindOne(ctx, map[string]string{"strVal": "val2"})
	val = &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
	coll.DeleteOne(ctx, map[string]string{"_id": val.ID.String()})
	cur = coll.FindOne(ctx, map[string]string{"_id": val.ID.String()})
	err = cur.Decode(val)
	if err != nil {
		fmt.Println(err)
	}

}

func TestMongocollctionFindOne(t *testing.T) {
	ctx := GetCorrelationContext()
	coll, err := mongo.NewDefaultCollection(ctx, MongoTestLogger, mongoURI, "GOLANGTEST", "Plain")
	if err != nil {
		t.Fatal(err)
	}
	cur := coll.FindOne(ctx, map[string]string{"strVal": "value1"})
	val := &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
}

func TestMongocollctionFindFetch(t *testing.T) {
	ctx := GetCorrelationContext()
	coll, err := mongo.NewDefaultCollection(ctx, MongoTestLogger, mongoURI, "GOLANGTEST", "Plain")
	if err != nil {
		t.Fatal(err)
	}
	loader := func(count int) []interface{} {
		val := make([]interface{}, count)
		for i := 0; i < count; i++ {
			val[i] = &TestVal{}
		}
		return val
	}
	data, err := coll.FindFetch(ctx, nil, loader)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range data {
		fmt.Printf("%+v\n", val)
	}
}

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

func TestMongocollectionPII(t *testing.T) {
	ctx := GetCorrelationContext()
	file, err := os.Open("piischeme.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	schemeByte, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	scheme := string(schemeByte)
	kmsArn := MongoTestConfig.KMS.Arn
	keyVaultNamespace := "encryption.__keyVault"
	keyAltName := "MongoPIITestKey"
	mongoURI := MongoTestConfig.Mongo.ConnectionString
	kmsProvider, err := mongo.GetDefaultAWSKMSProvider(ctx, MongoTestLogger, kmsArn)
	if err != nil {
		t.Fatal(err)
	}
	err = testutils.SetEncryptionKey(ctx, MongoTestLogger, &scheme, mongoURI, keyVaultNamespace, keyAltName, kmsProvider)
	if err != nil {
		t.Fatal(err)
	}
	coll, err := mongo.NewDefaultCollection(ctx, MongoTestLogger, mongoURI, "GOLANGTEST", "PII")
	piicoll, err := mongo.NewDefaultCSFLECollection(ctx, MongoTestLogger, mongoURI, keyVaultNamespace, "GOLANGTEST", "PII", scheme, kmsProvider, []string{"pan", "email"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = piicoll.InsertOneWithHash(ctx, map[string]interface{}{
		"dob":   "1991-08-02",
		"name":  "Vamshi Krishna",
		"pan":   "ABCDE1234F",
		"email": "vamshi@gosabariram.com",
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
	cur := piicoll.FindOneWithHash(ctx, map[string]interface{}{"email": "vamshi@gosabariram.com"})
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
		"email": "vamshi@gosabariram.com",
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
