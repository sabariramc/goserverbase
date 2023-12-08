package mongo

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"

	"github.com/sabariramc/goserverbase/v4/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Collection struct {
	*mongo.Collection
	log          *log.Logger
	hashFieldMap map[string]interface{}
}

var dec, _ = decimal.NewFromString("1.234")
var decimalType = reflect.TypeOf(dec)
var newRegistry = bson.NewRegistry()

func init() {
	newRegistry.RegisterTypeEncoder(decimalType, bsoncodec.ValueEncoderFunc(func(ctx bsoncodec.EncodeContext, vr bsonrw.ValueWriter, val reflect.Value) error {
		decimalValue, ok := val.Interface().(decimal.Decimal)
		if !ok {
			return fmt.Errorf("decimal bson encode error - value is not a decimal type - %v", val)
		}
		dec, err := primitive.ParseDecimal128(decimalValue.String())
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder : %w", err)
		}
		err = vr.WriteDecimal128(dec)
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder : %w", err)
		}
		return nil
	}))
	newRegistry.RegisterTypeDecoder(decimalType, bsoncodec.ValueDecoderFunc(func(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
		read, err := vr.ReadDecimal128()
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder : %w", err)
		}
		dec, err := decimal.NewFromString(read.String())
		if err != nil {
			return fmt.Errorf("mongo.DecimalRegistryBuilder : %w", err)
		}
		val.Set(reflect.ValueOf(dec))
		return nil
	}))
}

func newCustomBsonRegistry() *bsoncodec.Registry {
	return newRegistry
}

func (m *Collection) SetHashList(hasList []string) {
	m.hashFieldMap = make(map[string]interface{}, len(hasList))
	for _, val := range hasList {
		m.hashFieldMap[val] = nil
	}
}
