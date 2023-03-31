package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	client   *mongo.Client
	log      *log.Logger
	database *mongo.Database
	c        *Config
}

var ErrNoDocuments = mongo.ErrNoDocuments

func New(ctx context.Context, logger *log.Logger, c Config) (*Mongo, error) {
	var client *mongo.Client
	var err error
	client, err = NewMongoClient(ctx, logger, &c)
	if err != nil {
		return nil, fmt.Errorf("mongo.NewMongo : %w", err)
	}
	return &Mongo{client: client, log: logger, c: &c, database: client.Database(c.DatabaseName)}, nil
}

func NewMongoClient(ctx context.Context, logger *log.Logger, c *Config) (*mongo.Client, error) {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(c.ConnectionString)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMinPoolSize(c.MinConnectionPool)
	connectionOptions.SetMaxPoolSize(c.MaxConnectionPool)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 5)
	client, err := mongo.Connect(ctx, connectionOptions)
	if err != nil {
		logger.Error(ctx, "Error creating mongo connection", err)
		return nil, fmt.Errorf("mongo.NewMongoClient : %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(ctx, "Error pinging mongo server", err)
		return nil, fmt.Errorf("mongo.NewMongoClient : %w", err)
	}
	return client, nil
}

func (m *Mongo) GetClient() *mongo.Client {
	return m.client
}
func (m *Mongo) GetLogger() *log.Logger {
	return m.log
}
