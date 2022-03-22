package mongo

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sabariram.com/goserverbase/db/mongo"
)

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
	client, err := mongo.NewMongo(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
	if err != nil {
		t.Fatal(err)
	}
	coll := client.NewCollection("Plain")
	val1, _ := decimal.NewFromString("123.1232")
	val2, _ := decimal.NewFromString("123.1232")
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{
		"strVal":       "value1",
		"intVal":       123,
		"deciVal":      val1.Add(val2),
		"timeValUTC":   time.Now().UTC(),
		"timeValLocal": time.Now(),
	}
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
	// fmt.Printf("%+v\n", val)
	// coll.DeleteOne(ctx, map[string]string{"_id": val.ID.String()})
	// cur = coll.FindOne(ctx, map[string]string{"_id": val.ID.String()})
	// err = cur.Decode(val)
	// if err != nil {
	// 	fmt.Println(err)
	// }

}

func TestMongocollctionFindOne(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.NewMongo(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
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

func TestMongocollctionFindFetch(t *testing.T) {
	ctx := GetCorrelationContext()
	client, err := mongo.NewMongo(ctx, MongoTestLogger, *MongoTestConfig.Mongo)
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
	data, err := coll.FindFetch(ctx, nil, loader)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range data {
		fmt.Printf("%+v\n", val)
	}
}
