package mongo

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"

	"github.com/sabariramc/goserverbase/v6/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Collection extends mongo.Collection with additional type register for github.com/shopspring/decimal
type Collection struct {
	*mongo.Collection
	log log.Log
}

var dec, _ = decimal.NewFromString("1.234")
var decimalType = reflect.TypeOf(dec)
var newRegistry = bson.NewRegistry()

func init() {
	newRegistry.RegisterTypeEncoder(decimalType, bsoncodec.ValueEncoderFunc(func(ctx bsoncodec.EncodeContext, vr bsonrw.ValueWriter, val reflect.Value) error {
		decimalValue, ok := val.Interface().(decimal.Decimal)
		if !ok {
			return fmt.Errorf("mongo.DecimalRegistryBuilder: decimal bson encode error: value is not a decimal type - %v", val)
		}
		dec, err := primitive.ParseDecimal128(decimalValue.String())
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder.parse: %w", err)
		}
		err = vr.WriteDecimal128(dec)
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder.write: %w", err)
		}
		return nil
	}))
	newRegistry.RegisterTypeDecoder(decimalType, bsoncodec.ValueDecoderFunc(func(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
		read, err := vr.ReadDecimal128()
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder.read: %w", err)
		}
		dec, err := decimal.NewFromString(read.String())
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder.new: %w", err)
		}
		val.Set(reflect.ValueOf(dec))
		return nil
	}))
}

func NewCustomBsonRegistry() *bsoncodec.Registry {
	return newRegistry
}
