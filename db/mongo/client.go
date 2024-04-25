package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v5/log"
	"github.com/sabariramc/goserverbase/v5/utils"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Mongo struct {
	*mongo.Client
	log  log.Log
	name string
}

type Tracer interface {
	MongoDB() *event.CommandMonitor
}

var ErrNoDocuments = mongo.ErrNoDocuments

func NewWithAWSRoleAuth(ctx context.Context, logger log.Log, c Config, t Tracer, opts ...*options.ClientOptions) (*Mongo, error) {
	opts = append(opts, options.Client().SetAuth(options.Credential{
		AuthMechanism: "MONGODB-AWS",
	}))
	return New(ctx, logger, c, t, opts...)
}

func New(ctx context.Context, logger log.Log, c Config, t Tracer, opts ...*options.ClientOptions) (*Mongo, error) {
	var err error
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(c.ConnectionString)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMinPoolSize(5)
	connectionOptions.SetMaxPoolSize(10)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 5)
	connectionOptions.SetCompressors([]string{"snappy", "zlib", "zstd"})
	connectionOptions.SetAppName(c.ServiceName)
	connectionOptions.SetReadPreference(readpref.SecondaryPreferred())
	connectionOptions.SetWriteConcern(writeconcern.W1())
	if t != nil {
		connectionOptions.SetMonitor(t.MongoDB())
	}
	if c.EnableLog {
		mongoLogger := &MongoLogger{log: logger.NewResourceLogger("MongoInternalLog"), ctx: log.GetContextWithCorrelation(context.Background(), log.GetDefaultCorrelationParam("MongoInternal"))}
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
	if err != nil {
		return nil, fmt.Errorf("mongo.New: %w", err)
	}
	if c.Name == "" {
		c.Name = "MongoClient"
	}
	return &Mongo{Client: client, name: c.Name, log: logger.NewResourceLogger("MongoClient")}, nil
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
	return m.name
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
