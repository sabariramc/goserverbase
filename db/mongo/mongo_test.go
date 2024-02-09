package mongo_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/db/mongo"
	"github.com/sabariramc/goserverbase/v5/utils"
	"github.com/shopspring/decimal"
	"gotest.tools/assert"
)

type TestVal struct {
	mongo.BaseMongoModel `bson:",inline"`
	TestId               string          `bson:"testId"`
	IntVal               int64           `bson:"intVal"`
	DecimalVal           decimal.Decimal `bson:"decimalVal"`
	StrVal               string          `bson:"strVal"`
	BoolVal              bool            `bson:"boolVal"`
	TimeValUTC           time.Time       `bson:"timeValUTC"`
	TimeValLocal         time.Time       `bson:"timeValLocal"`
}

func GetSampleData() *TestVal {
	val1, _ := decimal.NewFromString("123.1232")
	val2, _ := decimal.NewFromString("123.1232")
	data := &TestVal{}
	data.TestId = utils.GenerateID(10, "test_")
	data.SetCreateParam("Random value")
	data.StrVal = "value1"
	data.IntVal = 123
	data.DecimalVal = val1.Add(val2)
	data.TimeValUTC = time.Now().Truncate(time.Second).UTC()
	data.TimeValLocal = time.Now().Truncate(time.Second).Local()
	return data
}

func BenchmarkMongoCollectionInsertOne(b *testing.B) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo, nil)
	assert.NilError(b, err)
	coll := client.Database("GOTEST").Collection("Plain")
	data := GetSampleData()
	for i := 0; i < b.N; i++ {
		_, err = coll.InsertOne(ctx, data)
		if err != nil {
			assert.NilError(b, err)
		}
	}
}

func TestMongoCollectionInsertOne(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo, nil)
	assert.NilError(t, err)
	coll := client.Database("GOTEST").Collection("Plain")
	data := GetSampleData()
	_, err = coll.InsertOne(ctx, data)
	assert.NilError(t, err)
}

func TestMongoCollection(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo, nil)
	assert.NilError(t, err)
	coll := client.Database("GOTEST").Collection("Plain")
	input := GetSampleData()
	coll.InsertOne(ctx, input)
	cur := coll.FindOne(ctx, map[string]string{"testId": input.TestId})
	res := &TestVal{}
	err = cur.Decode(res)
	assert.NilError(t, err)
	res.ID = nil
	input.BaseMongoDocument = nil
	res.BaseMongoDocument = nil
	assert.DeepEqual(t, res, input)
	coll.UpdateByID(ctx, res.ID, map[string]map[string]interface{}{"$set": {"strVal": "val2"}})
	cur = coll.FindOne(ctx, map[string]string{"testId": input.TestId})
	res = &TestVal{}
	err = cur.Decode(res)
	assert.NilError(t, err)
	coll.DeleteOne(ctx, map[string]string{"_id": res.ID.String()})
	cur = coll.FindOne(ctx, map[string]string{"_id": res.ID.String()})
	err = cur.Decode(res)
	fmt.Println(err)

}

func TestMongoCollectionFindOne(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo, nil)
	assert.NilError(t, err)
	coll := client.Database("GOTEST").Collection("Plain")
	cur := coll.FindOne(ctx, map[string]string{"strVal": "value1"})
	val := &TestVal{}
	err = cur.Decode(val)
	assert.NilError(t, err)
	fmt.Printf("%+v\n", val)
}
