package mongo

import (
	"github.com/sabariramc/goserverbase/v4/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
	log *log.Logger
}

var decimalHandlerOptions = options.Collection().SetRegistry(newCustomBsonRegistry())

func (d *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	opts = append(opts, decimalHandlerOptions)
	coll := d.Database.Collection(name, opts...)
	return &Collection{Collection: coll, log: d.log.NewResourceLogger("MongoCollection")}
}
