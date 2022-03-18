package mongo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
	"sabariram.com/goserverbase/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	databaseName   string
	collectionName string
	collection     *mongo.Collection
	ctx            context.Context
	log            *log.Log
	hashFieldMap   map[string]interface{}
}

var dec, _ = decimal.NewFromString("1.234")
var decimalType = reflect.TypeOf(dec)
var newRegistory = bson.NewRegistryBuilder().RegisterTypeEncoder(decimalType, bsoncodec.ValueEncoderFunc(func(ctx bsoncodec.EncodeContext, vr bsonrw.ValueWriter, val reflect.Value) error {
	custDec, ok := val.Interface().(decimal.Decimal)
	if !ok {
		return fmt.Errorf("decimal bson encode error - value is not a decimal type - %v", val)
	}
	dec, err := primitive.ParseDecimal128(custDec.String())
	if err != nil {
		return err
	}
	err = vr.WriteDecimal128(dec)
	if err != nil {
		return err
	}
	return nil
})).RegisterTypeDecoder(decimalType, bsoncodec.ValueDecoderFunc(func(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	read, err := vr.ReadDecimal128()
	if err != nil {
		return err
	}
	dec, err := decimal.NewFromString(read.String())
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(dec))
	return nil
}))

func NewDefaultCollection(ctx context.Context, mongoURI, databaseName, collectionName string) (*Collection, error) {
	logger := log.GetDefaultLogger()
	client, err := GetClient(ctx, mongoURI, logger)
	if err != nil {
		return nil, err
	}
	return NewCollection(ctx, client.client, databaseName, collectionName, logger), nil
}

func NewCollection(ctx context.Context, client *mongo.Client, databaseName, collectionName string, log *log.Log) *Collection {
	collectionOptions := options.Collection()
	collectionOptions.SetRegistry(newCustomBsonRegistory().Build())
	return &Collection{databaseName: databaseName, collectionName: collectionName, collection: client.Database(databaseName).Collection(collectionName, collectionOptions), ctx: ctx, log: log}
}

func newCustomBsonRegistory() *bsoncodec.RegistryBuilder {
	return newRegistory
}

func (m *Collection) SetHashList(hasList []string) {
	m.hashFieldMap = make(map[string]interface{}, len(hasList))
	for _, val := range hasList {
		m.hashFieldMap[val] = nil
	}
}
