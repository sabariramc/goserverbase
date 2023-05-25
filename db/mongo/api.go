package mongo

import (
	"context"
	"fmt"

	"github.com/sabariramc/goserverbase/v3/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NewLoadContainer func(count int) []interface{}

func (m *Collection) FindWithHash(ctx context.Context, filter map[string]interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	return m.Find(ctx, m.newHashFilter(ctx, filter), opts...)
}

func (m *Collection) FindOneWithHash(ctx context.Context, filter map[string]interface{}, opts ...*options.FindOneOptions) (s *mongo.SingleResult) {
	return m.FindOne(ctx, m.newHashFilter(ctx, filter), opts...)
}

func (m *Collection) FindFetch(ctx context.Context, loader NewLoadContainer, filter interface{}, opts ...*options.FindOptions) (res []interface{}, err error) {
	if filter == nil {
		filter = bson.D{}
	}
	cur, err := m.Find(ctx, filter, opts...)
	if err != nil {
		return
	}
	count := cur.RemainingBatchLength()
	valList := loader(count)
	i := 0
	for cur.Next(ctx) {
		err := cur.Decode(valList[i])
		if err != nil {
			return nil, fmt.Errorf("Collection.FindFetch : %w", err)
		}
		i++
	}
	res = valList
	m.log.Debug(ctx, "Mongo find fetch response", res)
	return
}

func (m *Collection) FindFetchWithHash(ctx context.Context, loader NewLoadContainer, filter map[string]interface{}, opts ...*options.FindOptions) (res []interface{}, err error) {
	return m.FindFetch(ctx, loader, m.newHashFilter(ctx, filter), opts...)
}

func (m *Collection) newHashFilter(ctx context.Context, filter map[string]interface{}) map[string]interface{} {
	if len(m.hashFieldMap) > 0 {
		hashFilter := make(map[string]interface{}, len(filter))
		for key, value := range filter {
			if _, ok := m.hashFieldMap[key]; ok {
				if strVal, ok := value.(string); ok {
					hashFilter[GetHashKey(key)] = utils.GetHash(strVal)
					continue
				}
				m.log.Warning(ctx, "Hash filter not generated for key - "+key, value)
			}
			hashFilter[key] = value
		}
		return hashFilter
	}
	return filter
}

func (m *Collection) InsertOneWithHash(ctx context.Context, doc map[string]interface{}, opts ...*options.InsertOneOptions) (ins interface{}, err error) {
	return m.InsertOne(ctx, m.newHashData(ctx, doc), opts...)

}

func (m *Collection) InsertManyWithHash(ctx context.Context, doc []map[string]interface{}, opts ...*options.InsertManyOptions) (ins interface{}, err error) {
	hashDoc := make([]interface{}, len(doc))
	for i, v := range doc {
		hashDoc[i] = m.newHashData(ctx, v)
	}
	return m.InsertMany(ctx, hashDoc, opts...)
}

func (m *Collection) UpdateByIDWithHash(ctx context.Context, id interface{}, update map[string]map[string]interface{}, opts ...*options.UpdateOptions) (upd *mongo.UpdateResult, err error) {
	update["$set"] = m.newHashData(ctx, update["$set"])
	return m.UpdateByID(ctx, id, update, opts...)
}

func (m *Collection) UpdateOneWithHash(ctx context.Context, filter interface{}, update map[string]map[string]interface{}, opts ...*options.UpdateOptions) (upd *mongo.UpdateResult, err error) {
	update["$set"] = m.newHashData(ctx, update["$set"])
	return m.UpdateOne(ctx, filter, update, opts...)

}

func (m *Collection) newHashData(ctx context.Context, data map[string]interface{}) map[string]interface{} {
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
					m.log.Warning(ctx, "Hash filter not generated for key - "+key, value)
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
