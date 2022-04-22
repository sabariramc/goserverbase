package mongo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"

	"github.com/sabariramc/goserverbase/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	collectionName string
	collection     *mongo.Collection
	log            *log.Logger
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
		return fmt.Errorf("mongo.DecimalRegistoryBuilder : %w", err)
	}
	err = vr.WriteDecimal128(dec)
	if err != nil {
		return fmt.Errorf("mongo.DecimalRegistoryBuilder : %w", err)
	}
	return nil
})).RegisterTypeDecoder(decimalType, bsoncodec.ValueDecoderFunc(func(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	read, err := vr.ReadDecimal128()
	if err != nil {
		return fmt.Errorf("mongo.DecimalRegistoryBuilder : %w", err)
	}
	dec, err := decimal.NewFromString(read.String())
	if err != nil {
		return fmt.Errorf("mongo.DecimalRegistoryBuilder : %w", err)
	}
	val.Set(reflect.ValueOf(dec))
	return nil
}))

var collectionOptions = options.Collection().SetRegistry(newCustomBsonRegistory().Build())

func (m *Mongo) NewCSFLECollection(ctx context.Context, collectionName string, hashFieldList []string) *Collection {
	if !m.isCSFLEEnabled {
		m.log.Emergency(ctx, "Non CSFLE Client", "Client passed is not a CSFLE Client", fmt.Errorf("CFLE COLLECTION ON NON CSFLE CLIENT"))
	}
	coll := &Collection{collectionName: collectionName, collection: m.database.Collection(collectionName, collectionOptions), log: m.log}
	coll.SetHashList(hashFieldList)
	return coll

}

func (m *Mongo) NewCollection(collectionName string) *Collection {
	return &Collection{collectionName: collectionName, collection: m.database.Collection(collectionName, collectionOptions), log: m.log}
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
