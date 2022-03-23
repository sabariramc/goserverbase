package mongo

import (
	"context"
	"fmt"
	"time"

	"sabariram.com/goserverbase/config"
	"sabariram.com/goserverbase/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	client         *mongo.Client
	log            *log.Logger
	database       *mongo.Database
	c              *config.MongoConfig
	csfle          *config.MongoCFLEConfig
	isCSFLEEnabled bool
}

var ErrNoDocuments = mongo.ErrNoDocuments

func NewMongo(ctx context.Context, logger *log.Logger, c config.MongoConfig) (*Mongo, error) {
	var client *mongo.Client
	var err error
	client, err = NewMongoClient(ctx, logger, &c)
	if err != nil {
		return nil, fmt.Errorf("mongo.NewMongo : %w", err)
	}
	return &Mongo{client: client, log: logger, c: &c, database: client.Database(c.DatabaseName), isCSFLEEnabled: false}, nil
}

func NewCSFLEMongo(client *mongo.Client, logger *log.Logger, c config.MongoConfig, csfle config.MongoCFLEConfig) *Mongo {
	return &Mongo{client: client, log: logger, c: &c, database: client.Database(c.DatabaseName), csfle: &csfle, isCSFLEEnabled: true}
}

func NewMongoClient(ctx context.Context, logger *log.Logger, c *config.MongoConfig) (*mongo.Client, error) {
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
