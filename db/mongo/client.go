package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/sabariramc/goserverbase/v4/log"
	"github.com/sabariramc/goserverbase/v4/utils"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	*mongo.Client
	log *log.Logger
}

var ErrNoDocuments = mongo.ErrNoDocuments

func NewWithAWSRoleAuth(ctx context.Context, logger *log.Logger, c Config, opts ...*options.ClientOptions) (*Mongo, error) {
	opts = append(opts, options.Client().SetAuth(options.Credential{
		AuthMechanism: "MONGODB-AWS",
	}))
	return New(ctx, logger, c, opts...)
}

func New(ctx context.Context, logger *log.Logger, c Config, opts ...*options.ClientOptions) (*Mongo, error) {
	var err error
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(c.ConnectionString)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMinPoolSize(c.MinConnectionPool)
	connectionOptions.SetMaxPoolSize(c.MaxConnectionPool)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 5)
	connectionOptions.SetCompressors([]string{"snappy", "zlib", "zstd"})
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
	opts = utils.Prepend[*options.ClientOptions](opts, connectionOptions)
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
	return NewWrapper(ctx, logger, client), nil
}

func NewWrapper(ctx context.Context, logger *log.Logger, client *mongo.Client) *Mongo {
	return &Mongo{Client: client, log: logger.NewResourceLogger("MongoClient")}
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

func (m *Mongo) Disconnect(ctx context.Context) error {
	m.log.Notice(ctx, "Mongo client closure initiated", nil)
	err := m.Client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("Mongo.Disconnect: %w", err)
	}
	m.log.Notice(ctx, "Mongo client closed", nil)
	return nil
}

type MongoLogger struct {
	log *log.Logger
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
