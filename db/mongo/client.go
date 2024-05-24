// Package mongo extends the functionality of [mongo]
// by adding tracing, logging, and implementing Shutdown and HealthCheck hooks.
package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/envvariables"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Mongo enhances mongo.Client with default option settings and logging capabilities.
type Mongo struct {
	*mongo.Client
	log        log.Log
	moduleName string
}

const ModuleName = "MongoClient"

// Tracer interface defines a method for obtaining a MongoDB command monitor.
type Tracer interface {
	MongoDB() *event.CommandMonitor
}

// GetDefaultConfig returns the default MongoDB client options including connection settings,
// application name, pool size, compression, read preference, write concern, logging, and monitoring.
func GetDefaultConfig(moduleName string, t Tracer, logger log.Log) *options.ClientOptions {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(utils.GetEnv(envvariables.MongoConnectionString, "mongodb://localhost:27017"))
	connectionOptions.SetAppName(utils.GetEnv(envvariables.ServiceName, "default"))
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMinPoolSize(5)
	connectionOptions.SetMaxPoolSize(10)
	connectionOptions.SetMaxConnIdleTime(5 * time.Minute)
	connectionOptions.SetCompressors([]string{"snappy", "zlib", "zstd"})
	connectionOptions.SetReadPreference(readpref.SecondaryPreferred())
	connectionOptions.SetWriteConcern(writeconcern.W1())
	if t != nil {
		connectionOptions.SetMonitor(t.MongoDB())
	}
	mongoLogger := &MongoLogger{
		log: logger.NewResourceLogger(moduleName + ":InternalLog"),
		ctx: correlation.GetContextWithCorrelationParam(context.Background(), correlation.NewCorrelationParam(moduleName+"-InternalLog")),
	}
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
	return connectionOptions
}

// NewWithDefaultOptions creates a new Mongo client with the default configuration options,
// incorporating additional options provided by the caller.
func NewWithDefaultOptions(logger log.Log, t Tracer, opts ...*options.ClientOptions) (*Mongo, error) {
	logger = logger.NewResourceLogger(ModuleName)
	connectionOptions := GetDefaultConfig(ModuleName, t, logger)
	opts = utils.Prepend(opts, connectionOptions)
	client, err := New(logger, opts...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// New creates a new Mongo client with the provided client options and initializes the connection.
func New(logger log.Log, opts ...*options.ClientOptions) (*Mongo, error) {
	ctx := correlation.GetContextWithCorrelationParam(context.Background(), &correlation.CorrelationParam{CorrelationID: ModuleName})
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		logger.Error(ctx, "error creating mongo connection", err)
		return nil, fmt.Errorf("mongo.New: error creating mongo connection: %w", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		logger.Error(ctx, "error pinging mongo server", err)
		return nil, fmt.Errorf("mongo.New: error pinging mongo server: %w", err)
	}
	return &Mongo{Client: client, moduleName: ModuleName, log: logger}, nil
}

// GetClient returns the underlying mongo.Client.
func (m *Mongo) GetClient() *mongo.Client {
	return m.Client
}

// GetLogger returns the logger associated with the Mongo client.
func (m *Mongo) GetLogger() log.Log {
	return m.log
}

// Database returns a wrapped mongo.Database with enhanced logging.
func (m *Mongo) Database(name string, opts ...*options.DatabaseOptions) *Database {
	db := m.Client.Database(name, opts...)
	return &Database{Database: db, log: m.log.NewResourceLogger("MongoDatabase:" + name)}
}

// Shutdown gracefully closes the Mongo client connection.
func (m *Mongo) Shutdown(ctx context.Context) error {
	m.log.Notice(ctx, "Mongo client closure initiated", nil)
	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("Mongo.Disconnect: %w", err)
	}
	m.log.Notice(ctx, "Mongo client closed", nil)
	return nil
}

// Name returns the module name.
func (m *Mongo) Name(ctx context.Context) string {
	return ModuleName
}

// SetModuleName sets the module name and updates the logger.
func (m *Mongo) SetModuleName(name string) {
	m.moduleName = name
	m.log = m.log.NewResourceLogger(name)
}

// HealthCheck performs a health check by pinging the MongoDB server.
func (m *Mongo) HealthCheck(ctx context.Context) error {
	return m.Ping(ctx, nil)
}

// MongoLogger is used for logging MongoDB events.
type MongoLogger struct {
	log log.Log
	ctx context.Context
}

// Info logs informational messages.
func (m *MongoLogger) Info(level int, message string, keysAndValues ...interface{}) {
	if level == int(options.LogLevelInfo) {
		m.log.Debug(m.ctx, message, keysAndValues)
	} else {
		m.log.Trace(m.ctx, message, keysAndValues)
	}
}

// Error logs error messages.
func (m *MongoLogger) Error(err error, message string, keysAndValues ...interface{}) {
	m.log.Error(m.ctx, message, keysAndValues)
}

// PoolEvent logs MongoDB pool events.
func (m *MongoLogger) PoolEvent(e *event.PoolEvent) {
	m.log.Debug(m.ctx, "mongo pool event", e)
}
