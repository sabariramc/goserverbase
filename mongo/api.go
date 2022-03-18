package mongo

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *Collection) Find(filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	xray.Capture(m.ctx, "MongoFind", func(ctx context.Context) error {
		cur, err = m.collection.Find(m.ctx, filter, opts...)
		if err != nil {
			m.log.Error("Mongo find error", err)
			xray.AddError(m.ctx, err)
		}
		return nil
	})
	return

}

func (m *Collection) FindOne(filter interface{}, opts ...*options.FindOneOptions) (s *mongo.SingleResult) {
	xray.Capture(m.ctx, "MongoFindOne", func(ctx1 context.Context) error {
		s = m.collection.FindOne(m.ctx, filter, opts...)
		return nil
	})
	return
}

type NewLoadContainer func(count int) []interface{}

func (m *Collection) FindFetch(filter interface{}, loader NewLoadContainer, opts ...*options.FindOptions) (res []interface{}, err error) {
	xray.Capture(m.ctx, "MongoFindFetch", func(ctx context.Context) error {
		if filter == nil {
			filter = bson.D{}
		}
		cur, err := m.Find(filter, opts...)
		if err != nil {
			return err
		}
		count := cur.RemainingBatchLength()
		valList := loader(count)
		i := 0
		for cur.Next(m.ctx) {
			cur.Decode(valList[i])
			i++
		}
		res = valList
		m.log.Debug("Mongo find fetch response", res)
		return nil
	})
	return
}

func (m *Collection) InsertOne(doc interface{}, opts ...*options.InsertOneOptions) (ins *mongo.InsertOneResult, err error) {
	xray.Capture(m.ctx, "MongoInsertOne", func(ctx1 context.Context) error {
		ins, err = m.collection.InsertOne(m.ctx, doc, opts...)
		m.log.Debug("Mongo insert one response", ins)
		return nil
	})
	return
}

func (m *Collection) InsertMany(doc []interface{}, opts ...*options.InsertManyOptions) (ins *mongo.InsertManyResult, err error) {
	xray.Capture(m.ctx, "MongoInsertMany", func(ctx1 context.Context) error {
		ins, err = m.collection.InsertMany(m.ctx, doc, opts...)
		if err != nil {
			m.log.Error("Mongo insert many error", err)
			return err
		}
		m.log.Debug("Mongo insert many response", ins)
		return nil
	})
	return
}

func (m *Collection) UpdateByID(id interface{}, update interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	xray.Capture(m.ctx, "MongoUpdateById", func(ctx1 context.Context) error {
		val, ok := id.(string)
		if ok {
			id, err = primitive.ObjectIDFromHex(val)
			if err != nil {
				m.log.Error("Mongo update by id, id creation error", err)
				return err
			}
		}
		res, err = m.collection.UpdateByID(m.ctx, id, update, opts...)
		if err != nil {
			m.log.Error("Mongo update by id error", err)
			return err
		}
		m.log.Debug("Mongo update by id response", res)
		return nil
	})
	return
}

func (m *Collection) UpdateMany(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	xray.Capture(m.ctx, "MongoUpdateMany", func(ctx1 context.Context) error {
		res, err = m.collection.UpdateMany(m.ctx, filter, update, opts...)
		if err != nil {
			m.log.Error("Mongo update many error", err)
			return err
		}
		m.log.Debug("Mongo update many response", res)
		return nil
	})
	return
}

func (m *Collection) UpdateOne(filter interface{}, update interface{}, opts ...*options.UpdateOptions) (res *mongo.UpdateResult, err error) {
	xray.Capture(m.ctx, "MongoUpdateOne", func(ctx1 context.Context) error {
		res, err = m.collection.UpdateOne(m.ctx, filter, update, opts...)
		if err != nil {
			m.log.Error("Mongo update one error", err)
			return err
		}
		m.log.Debug("Mongo update one response", res)
		return nil
	})
	return
}

func (m *Collection) DeleteOne(filter interface{}, opts ...*options.DeleteOptions) (res *mongo.DeleteResult, err error) {
	xray.Capture(m.ctx, "MongoDeleteOne", func(ctx1 context.Context) error {
		res, err = m.collection.DeleteOne(m.ctx, filter, opts...)
		if err != nil {
			m.log.Error("Mongo delete one error", err)
			return err
		}
		m.log.Debug("Mongo delete one response", res)
		return nil
	})
	return
}

func (m *Collection) DeleteMany(filter interface{}, opts ...*options.DeleteOptions) (res *mongo.DeleteResult, err error) {
	xray.Capture(m.ctx, "MongoDeleteMany", func(ctx1 context.Context) error {
		res, err = m.collection.DeleteMany(m.ctx, filter, opts...)
		if err != nil {
			m.log.Error("Mongo delete one error", err)
			return err
		}
		m.log.Debug("Mongo delete many response", res)
		return nil
	})
	return
}
