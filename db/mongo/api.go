package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Collection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	return m.collection.Find(ctx, filter, opts...)

}

func (m *Collection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (s *mongo.SingleResult) {
	return m.collection.FindOne(ctx, filter, opts...)
}

type NewLoadContainer func(count int) []interface{}

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
		cur.Decode(valList[i])
		i++
	}
	res = valList
	m.log.Debug(ctx, "Mongo find fetch response", res)
	return
}

func (m *Collection) InsertOne(ctx context.Context, doc interface{}, opts ...*options.InsertOneOptions) (ins *mongo.InsertOneResult, err error) {
	ins, err = m.collection.InsertOne(ctx, doc, opts...)
	m.log.Debug(ctx, "Mongo insert one response", ins)
	return
}

func (m *Collection) InsertMany(ctx context.Context, doc []interface{}, opts ...*options.InsertManyOptions) (ins *mongo.InsertManyResult, err error) {
	ins, err = m.collection.InsertMany(ctx, doc, opts...)
	if err != nil {
		m.log.Error(ctx, "Mongo insert many error", err)
		return
	}
	m.log.Debug(ctx, "Mongo insert many response", ins)
	return
}

func (m *Collection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	val, ok := id.(string)
	if ok {
		id, err = primitive.ObjectIDFromHex(val)
		if err != nil {
			m.log.Error(ctx, "Mongo update by id, id creation error", err)
			return
		}
	}
	res, err = m.collection.UpdateByID(ctx, id, update, opts...)
	if err != nil {
		m.log.Error(ctx, "Mongo update by id error", err)
		return
	}
	m.log.Debug(ctx, "Mongo update by id response", res)
	return
}

func (m *Collection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = m.collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		m.log.Error(ctx, "Mongo update many error", err)
		return
	}
	m.log.Debug(ctx, "Mongo update many response", res)
	return
}

func (m *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	res, err = m.collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		m.log.Error(ctx, "Mongo update one error", err)
		return
	}
	m.log.Debug(ctx, "Mongo update one response", res)
	return
}

func (m *Collection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (res *mongo.DeleteResult, err error) {
	res, err = m.collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		m.log.Error(ctx, "Mongo delete one error", err)
		return
	}
	m.log.Debug(ctx, "Mongo delete one response", res)
	return
}

func (m *Collection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (res *mongo.DeleteResult, err error) {
	res, err = m.collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		m.log.Error(ctx, "Mongo delete one error", err)
		return
	}
	m.log.Debug(ctx, "Mongo delete many response", res)
	return
}
