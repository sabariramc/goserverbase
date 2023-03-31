package mongo_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/db/mongo"
	"github.com/shopspring/decimal"
)

type TestVal struct {
	mongo.BaseMongoModel `bson:",inline"`
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
	data.SetCreateParam("Random value")
	data.StrVal = "value1"
	data.IntVal = 123
	data.DecimalVal = val1.Add(val2)
	data.TimeValUTC = time.Now().UTC()
	data.TimeValLocal = time.Now()
	return data
}

func TestMongoCollectionInsertOne(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		t.Fatal(err)
	}
	coll := client.NewCollection("Plain")
	data := GetSampleData()
	fmt.Printf("%+v\n", data)
	_, err = coll.InsertOne(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
	data = GetSampleData()
	_, err = coll.InsertOne(ctx, data)
}

func TestMongoCollection(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		t.Fatal(err)
	}
	coll := client.NewCollection("Plain")
	data := GetSampleData()
	fmt.Printf("%+v\n", data)
	coll.InsertOne(ctx, data)
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

func TestMongoCollectionFindOne(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		t.Fatal(err)
	}
	coll := client.NewCollection("Plain")
	cur := coll.FindOne(ctx, map[string]string{"strVal": "val2"})
	val := &TestVal{}
	err = cur.Decode(val)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", val)
}

func TestMongoCollectionFindFetch(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.New(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		t.Fatal(err)
	}
	coll := client.NewCollection("Plain")
	loader := func(count int) []interface{} {
		val := make([]interface{}, count)
		for i := 0; i < count; i++ {
			val[i] = &TestVal{}
		}
		return val
	}
	data, err := coll.FindFetch(ctx, loader, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range data {
		fmt.Printf("%+v\n", val)
	}
}
