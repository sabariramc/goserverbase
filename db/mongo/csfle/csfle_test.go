package csfle_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/aws"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/db/mongo/csfle"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type Name struct {
	First  string `bson:"first"`
	Middle string `bson:"middle"`
	Last   string `bson:"last"`
	Full   string `bson:"full"`
}

type PIITestVal struct {
	ID      primitive.ObjectID `bson:"_id"`
	DOB     time.Time          `bson:"dob"`
	Name    Name               `bson:"name"`
	Pan     string             `bson:"pan"`
	Email   string             `bson:"email"`
	Phone   []string           `bson:"phone"`
	Address Address            `bson:"address"`
}

func getKMSProvider(ctx context.Context, kmsArn string) (csfle.MasterKeyProvider, error) {
	if os.Getenv("AWS_PROVIDER") == "local" {
		cred, _ := aws.GetDefaultAWSConfig().Credentials.Retrieve(ctx)
		kmsProvider := csfle.NewAWSProvider(map[string]interface{}{
			"accessKeyId":     cred.AccessKeyID,
			"secretAccessKey": cred.SecretAccessKey,
		}, csfle.AWSDataKeyOpts{
			Region:   aws.GetDefaultAWSConfig().Region,
			KeyARN:   kmsArn,
			Endpoint: "http://localhost.localstack.cloud:4566",
		})
		return kmsProvider, nil
	}
	return csfle.GetAWSMasterKeyProvider(ctx, MongoTestLogger, kmsArn)
}

func TestCollectionPII(t *testing.T) {
	ctx := GetCorrelationContext()
	file, err := os.Open("./sample/piischeme.json")
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, file.Close())
	}()
	schemeByte, err := io.ReadAll(file)
	assert.NilError(t, err)
	scheme := string(schemeByte)
	kmsArn := MongoTestConfig.AWS.KMS_ARN
	dbName, collName := "GOTEST", "PII"
	kmsProvider, err := getKMSProvider(ctx, kmsArn)
	assert.NilError(t, err)
	config := MongoTestConfig.CSFLE
	dbScheme, err := csfle.SetEncryptionKey(ctx, MongoTestLogger, &scheme, *MongoTestConfig.Mongo, config.KeyVaultNamespace, kmsProvider)
	assert.NilError(t, err)
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	assert.NilError(t, err)

	config.KMSCredentials = kmsProvider.Credentials()
	config.SchemaMap = dbScheme
	csfleClient, err := csfle.New(ctx, MongoTestLogger, *config)
	assert.NilError(t, err)
	// csfleClient.Database(dbName).Collection(collName).Drop(context.TODO())
	// assert.NilError(t, err)
	// err = csfleClient.Database(dbName).CreateCollection(context.TODO(), collName)
	// assert.NilError(t, err)
	piicoll := csfleClient.Database(dbName).Collection(collName, options.Collection())
	uuid := "FAsdfasfsadfsdafs"
	dob, err := time.Parse(time.DateOnly, "1991-08-02")
	coll := client.Database(dbName).Collection(collName)
	assert.NilError(t, err)
	_, err = piicoll.InsertOne(ctx, map[string]interface{}{
		"dob": dob,
		"name": map[string]any{
			"first":  "first name person 1",
			"middle": "middle name person 1",
			"last":   "last name person 1",
			"full":   "full name person 1",
		},
		"pan":   "ABCDE1234F",
		"email": "sab@sabariram.com",
		"address": map[string]string{
			"addressLine1": "door no with street name",
			"addressLine2": "taluk and postal office",
			"addressLine3": "Optional landmark",
			"state":        "TEST",
			"pin":          "TEST",
			"country":      "India",
		},
		"UUID": uuid,
	})
	assert.NilError(t, err)
	cur := piicoll.FindOne(ctx, map[string]interface{}{"UUID": uuid})
	val := &PIITestVal{}
	err = cur.Decode(val)
	assert.NilError(t, err)
	fmt.Printf("%+v\n", val)
	piicoll.UpdateOne(ctx, val.ID, map[string]map[string]interface{}{"$set": {"UUID": uuid}})
	cur = piicoll.FindOne(ctx, map[string]interface{}{"UUID": uuid})
	val = &PIITestVal{}
	err = cur.Decode(val)
	assert.NilError(t, err)
	fmt.Printf("%+v\n", val)
	cur = piicoll.FindOne(ctx, map[string]interface{}{"_id": val.ID})
	err = cur.Decode(val)
	assert.NilError(t, err)
	data, err := coll.Find(ctx, map[string]map[string]interface{}{"pan": {"$exists": true}})
	assert.NilError(t, err)
	for data.Next(ctx) {
		decodeData := make(map[string]interface{})
		data.Decode(&decodeData)
		fmt.Printf("%+v\n", decodeData)
	}
	res, err := piicoll.DeleteOne(ctx, map[string]interface{}{"_id": val.ID})
	assert.NilError(t, err)
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
		"dob": dob,
		"name": map[string]any{
			"first":  "first name person 2",
			"middle": "middle name person 2",
			"last":   "last name person 2",
			"full":   "full name person 2",
		},
		"pan":   "ABCDE1234F",
		"email": "sab@sabariram.com",
		"address": map[string]string{
			"addressLine1": "door no with street name",
			"addressLine2": "taluk and postal office",
			"addressLine3": "Optional landmark",
			"state":        "TEST",
			"pin":          "TEST",
			"country":      "India",
		},
		"UUID": "FAsdfasfsadfsdafs",
	})
	assert.NilError(t, err)
}
