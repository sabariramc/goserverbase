package mongo

import (
	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	*mongo.Database
	log *log.Logger
}

var customRegistryOption = options.Collection().SetRegistry(NewCustomBsonRegistry())

func (d *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	opts = utils.Prepend(opts, customRegistryOption)
	coll := d.Database.Collection(name, opts...)
	return &Collection{Collection: coll, log: d.log.NewResourceLogger("MongoCollection:" + name)}
}
