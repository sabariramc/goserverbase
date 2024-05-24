package mongo

import (
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database extends mongo.Database by incorporating enhanced logging functionality.
type Database struct {
	*mongo.Database
	log log.Log
}

// customRegistryOption is a collection option that sets a custom BSON registry to handle decimal.Decimal types.
var customRegistryOption = options.Collection().SetRegistry(NewCustomBsonRegistry())

// Collection returns a Collection with a custom BSON registry and enhanced logging.
// It prepends the custom registry option to any other collection options provided.
func (d *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	opts = utils.Prepend(opts, customRegistryOption)
	coll := d.Database.Collection(name, opts...)
	return &Collection{Collection: coll, log: d.log.NewResourceLogger("MongoCollection:" + name)}
}
