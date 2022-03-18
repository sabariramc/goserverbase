package mongo

import (
	"context"
	"fmt"

	"github.com/aws/aws-xray-sdk-go/xray"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sabariram.com/goserverbase/utils"
)

func (m *Collection) FindWithHash(filter map[string]interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	xray.Capture(m.ctx, "MongoFindWithHash", func(c context.Context) error {
		cur, err = m.Find(m.newHashFilter(filter), opts...)
		return err
	})
	return
}

func (m *Collection) FindOneWithHash(filter map[string]interface{}, opts ...*options.FindOneOptions) (s *mongo.SingleResult) {
	xray.Capture(m.ctx, "MongoFindOneWithHash", func(c context.Context) error {
		s = m.FindOne(m.newHashFilter(filter), opts...)
		return nil
	})
	return
}

func (m *Collection) FindFetchWithHash(filter map[string]interface{}, loader NewLoadContainer, opts ...*options.FindOptions) (res []interface{}, err error) {
	xray.Capture(m.ctx, "MongoFindWithHash", func(c context.Context) error {
		res, err = m.FindFetch(m.newHashFilter(filter), loader, opts...)
		return err
	})
	return
}

func (m *Collection) newHashFilter(filter map[string]interface{}) map[string]interface{} {
	if len(m.hashFieldMap) > 0 {
		hashFilter := make(map[string]interface{}, len(filter))
		for key, value := range filter {
			if _, ok := m.hashFieldMap[key]; ok {
				if strVal, ok := value.(string); ok {
					hashFilter[GetHashKey(key)] = utils.GetHash(strVal)
					continue
				}
				m.log.Warning("Hash filter not gererate for key - "+key, value)
			}
			hashFilter[key] = value
		}
		return hashFilter
	}
	return filter
}

func (m *Collection) InsertOneWithHash(doc map[string]interface{}, opts ...*options.InsertOneOptions) (ins interface{}, err error) {
	xray.Capture(m.ctx, "MongoInsertOneWithHash", func(c context.Context) error {
		ins, err = m.InsertOne(m.newHashData(doc), opts...)
		return err
	})
	return
}

func (m *Collection) InsertManyWithHash(doc []map[string]interface{}, opts ...*options.InsertManyOptions) (ins interface{}, err error) {

	xray.Capture(m.ctx, "MongoInsertManyWithHash", func(c context.Context) error {
		hashDoc := make([]interface{}, len(doc))
		for i, v := range doc {
			hashDoc[i] = m.newHashData(v)
		}
		ins, err = m.InsertMany(hashDoc, opts...)
		return err
	})
	return
}

func (m *Collection) UpdateByIDWithHash(id interface{}, update map[string]map[string]interface{}, opts ...*options.UpdateOptions) (upd *mongo.UpdateResult, err error) {
	xray.Capture(m.ctx, "MongoUpdateByIDWithHash", func(c context.Context) error {
		update["$set"] = m.newHashData(update["$set"])
		upd, err = m.UpdateByID(id, update, opts...)
		return err
	})
	return
}

func (m *Collection) UpdateOneWithHash(filter interface{}, update map[string]map[string]interface{}, opts ...*options.UpdateOptions) (upd *mongo.UpdateResult, err error) {
	xray.Capture(m.ctx, "MongoUpdateByIDWithHash", func(c context.Context) error {
		update["$set"] = m.newHashData(update["$set"])
		upd, err = m.UpdateOne(filter, update, opts...)
		return err
	})
	return

}

func (m *Collection) newHashData(data map[string]interface{}) map[string]interface{} {
	if len(m.hashFieldMap) > 0 {
		hashData := make(map[string]interface{}, len(data))
		for key, value := range data {
			if _, ok := m.hashFieldMap[key]; ok {
				switch v := value.(type) {
				case string:
					hashData[GetHashKey(key)] = utils.GetHash(v)
				case []string:
					hashList := make([]string, len(v))
					for i, lv := range v {
						hashList[i] = utils.GetHash(lv)
					}
					hashData[GetHashKey(key)] = hashList
				default:
					m.log.Warning("Hash filter not gererate for key - "+key, value)
				}
			}
			hashData[key] = value
		}
		return hashData
	}
	return data
}

func GetHashKey(key string) string {
	return fmt.Sprintf("%vHash", key)
}
