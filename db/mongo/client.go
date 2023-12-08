package mongo

import (
	"context"
	"fmt"
	"time"

	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"

	"github.com/sabariramc/goserverbase/v4/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	*mongo.Client
	log *log.Logger
	c   *Config
}

var ErrNoDocuments = mongo.ErrNoDocuments

func NewWithAWSRoleAuth(ctx context.Context, logger *log.Logger, c Config, opts ...*options.ClientOptions) (*Mongo, error) {
	opts = append(opts, options.Client().SetAuth(options.Credential{
		AuthMechanism: "MONGODB-AWS",
	}))
	return New(ctx, logger, c, opts...)
}

func New(ctx context.Context, logger *log.Logger, c Config, opts ...*options.ClientOptions) (*Mongo, error) {
	var client *mongo.Client
	var err error
	mon := options.Client()
	mon.Monitor = mongotrace.NewMonitor()
	opts = append(opts, mon)
	client, err = NewMongoClient(ctx, logger, &c, opts...)
	if err != nil {
		return nil, fmt.Errorf("mongo.NewMongo: %w", err)
	}
	return NewWrapper(ctx, logger, c, client), nil
}

func NewWrapper(ctx context.Context, logger *log.Logger, c Config, client *mongo.Client) *Mongo {
	return &Mongo{Client: client, log: logger.NewResourceLogger("MongoClient"), c: &c}
}

func NewMongoClient(ctx context.Context, logger *log.Logger, c *Config, opts ...*options.ClientOptions) (*mongo.Client, error) {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(c.ConnectionString)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMinPoolSize(c.MinConnectionPool)
	connectionOptions.SetMaxPoolSize(c.MaxConnectionPool)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 5)
	opts = append(opts, connectionOptions)
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		logger.Error(ctx, "error creating mongo connection", err)
		return nil, fmt.Errorf("mongo.NewMongoClient: error creating mongo connection: %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(ctx, "error pinging mongo server", err)
		return nil, fmt.Errorf("mongo.NewMongoClient: error pinging mongo server: %w", err)
	}
	return client, nil
}

func (m *Mongo) GetClient() *mongo.Client {
	return m.Client
}
func (m *Mongo) GetLogger() *log.Logger {
	return m.log
}

func (m *Mongo) Database(name string, opts ...*options.DatabaseOptions) *Database {
	db := m.Client.Database(name, opts...)
	return &Database{Database: db, log: m.log.NewResourceLogger("MongoDatabase")}
}
