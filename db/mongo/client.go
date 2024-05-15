// Package mongo extends the functionality of go.mongodb.org/mongo-driver/mongo with tracing, logging and implements Shutdown and HealthCheck hooks
package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Mongo extends mongo.Client with default options setting and logging. Enable mongo internal log with a flag
type Mongo struct {
	*mongo.Client
	log        log.Log
	moduleName string
}

type Tracer interface {
	MongoDB() *event.CommandMonitor
}

func NewWithDefaultOptions(ctx context.Context, serviceName string, logger log.Log, c Config, t Tracer, opts ...*options.ClientOptions) (*Mongo, error) {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(c.ConnectionString)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMinPoolSize(5)
	connectionOptions.SetMaxPoolSize(10)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 5)
	connectionOptions.SetCompressors([]string{"snappy", "zlib", "zstd"})
	connectionOptions.SetAppName(serviceName)
	connectionOptions.SetReadPreference(readpref.SecondaryPreferred())
	connectionOptions.SetWriteConcern(writeconcern.W1())
	if t != nil {
		connectionOptions.SetMonitor(t.MongoDB())
	}
	if c.EnableLog {
		mongoLogger := &MongoLogger{log: logger.NewResourceLogger("MongoInternalLog"), ctx: log.GetContextWithCorrelationParam(context.Background(), log.GetDefaultCorrelationParam("MongoInternal"))}
		connectionOptions.SetLoggerOptions(&options.LoggerOptions{
			ComponentLevels: map[options.LogComponent]options.LogLevel{
				options.LogComponentAll: options.LogLevelDebug,
			},
			Sink:              mongoLogger,
			MaxDocumentLength: 1024,
		})
		connectionOptions.SetPoolMonitor(&event.PoolMonitor{
			Event: mongoLogger.PoolEvent,
		})
	}
	opts = utils.Prepend(opts, connectionOptions)
	client, err := New(ctx, logger, opts...)
	if err != nil {
		return nil, err
	}
	if c.ModuleName != "" {
		client.moduleName = c.ModuleName
	}
	return client, nil
}

func New(ctx context.Context, logger log.Log, opts ...*options.ClientOptions) (*Mongo, error) {
	var err error
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		logger.Error(ctx, "error creating mongo connection", err)
		return nil, fmt.Errorf("mongo.New: error creating mongo connection: %w", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(ctx, "error pinging mongo server", err)
		return nil, fmt.Errorf("mongo.New: error pinging mongo server: %w", err)
	}
	return &Mongo{Client: client, moduleName: "MongoClient", log: logger.NewResourceLogger("MongoClient")}, nil
}

func (m *Mongo) GetClient() *mongo.Client {
	return m.Client
}
func (m *Mongo) GetLogger() log.Log {
	return m.log
}

func (m *Mongo) Database(name string, opts ...*options.DatabaseOptions) *Database {
	db := m.Client.Database(name, opts...)
	return &Database{Database: db, log: m.log.NewResourceLogger("MongoDatabase")}
}

func (m *Mongo) Shutdown(ctx context.Context) error {
	m.log.Notice(ctx, "Mongo client closure initiated", nil)
	err := m.Client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("Mongo.Disconnect: %w", err)
	}
	m.log.Notice(ctx, "Mongo client closed", nil)
	return nil
}

func (m *Mongo) Name(ctx context.Context) string {
	if m.moduleName == "" {
		return "MongoClient"
	}
	return m.moduleName
}

func (m *Mongo) HealthCheck(ctx context.Context) error {
	return m.Ping(ctx, nil)
}

type MongoLogger struct {
	log log.Log
	ctx context.Context
}

func (m *MongoLogger) Info(level int, message string, keysAndValues ...interface{}) {
	if level == int(options.LogLevelInfo) {
		m.log.Debug(m.ctx, message, keysAndValues)
	} else {
		m.log.Trace(m.ctx, message, keysAndValues)
	}
}

func (m *MongoLogger) Error(err error, message string, keysAndValues ...interface{}) {
	m.log.Error(m.ctx, message, keysAndValues)
}

func (m *MongoLogger) PoolEvent(e *event.PoolEvent) {
	m.log.Debug(m.ctx, "mongo pool event", e)
}
