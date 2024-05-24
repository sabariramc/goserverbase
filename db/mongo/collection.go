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

// Collection extends mongo.Collection by adding support for encoding and decoding [decimal] types.
type Collection struct {
	*mongo.Collection
	log log.Log
}

var (
	dec         = decimal.New(1, 3)   // Example decimal value used to get the decimal type.
	decimalType = reflect.TypeOf(dec) // Type representation of the decimal.Decimal type.
	newRegistry = bson.NewRegistry()  // Custom BSON registry for decimal type.
)

// init initializes the custom BSON registry by registering encoders and decoders for the decimal.Decimal type.
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
		if err = vr.WriteDecimal128(dec); err != nil {
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

// NewCustomBsonRegistry returns the custom BSON registry configured to handle decimal.Decimal types.
func NewCustomBsonRegistry() *bsoncodec.Registry {
	return newRegistry
}
