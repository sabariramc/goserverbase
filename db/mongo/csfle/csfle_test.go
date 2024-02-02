package csfle_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v5/db/mongo/csfle/sample"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gotest.tools/assert"
)

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
	kmsProvider, err := sample.GetKMSProvider(ctx, MongoTestLogger, kmsArn)
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
	coll := client.Database(dbName).Collection(collName)
	assert.NilError(t, err)
	uuid1 := uuid.New().String()
	data1 := sample.GetRandomData(uuid1)
	_, err = piicoll.InsertOne(ctx, data1)
	assert.NilError(t, err)
	cur := piicoll.FindOne(ctx, map[string]interface{}{"UUID": uuid1})
	val := sample.PIITestVal{}
	err = cur.Decode(&val)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, data1)
	data1.Pan = "ABCDE1235F"
	piicoll.UpdateOne(ctx, map[string]string{
		"UUID": uuid1,
	}, map[string]map[string]interface{}{"$set": {"pan": "ABCDE1235F"}})
	cur = piicoll.FindOne(ctx, map[string]interface{}{"UUID": uuid1})
	val = sample.PIITestVal{}
	err = cur.Decode(&val)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, data1)
	_, err = piicoll.InsertOne(ctx, sample.GetRandomData(uuid1))
	assert.NilError(t, err)
	cur = coll.FindOne(ctx, map[string]interface{}{"UUID": uuid1})
	decodeData := &map[string]any{}
	err = cur.Decode(decodeData)
	assert.NilError(t, err)
	fmt.Print(decodeData)
	res, err := piicoll.DeleteMany(ctx, map[string]interface{}{"UUID": uuid1})
	assert.NilError(t, err)
	assert.Equal(t, 2, int(res.DeletedCount))
}
