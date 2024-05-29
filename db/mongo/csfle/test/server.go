package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
	"github.com/sabariramc/goserverbase/v6/correlation"
	"github.com/sabariramc/goserverbase/v6/db/mongo"
	"github.com/sabariramc/goserverbase/v6/db/mongo/csfle"
	"github.com/sabariramc/goserverbase/v6/db/mongo/csfle/sample"
	"github.com/sabariramc/goserverbase/v6/instrumentation"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/testutils"
	"github.com/sabariramc/goserverbase/v6/utils"
)

var ServerTestConfig *testutils.TestConfig
var ServerTestLogger log.Log
var ServerTestLMux log.Mux

const ServiceName = "BaseTest"

func init() {
	fmt.Println(os.Getwd())
	testutils.LoadEnv("../../../.env")
	ServerTestConfig = testutils.NewConfig()
	ServerTestLogger = log.New(log.WithServiceName(ServiceName))
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}

type server struct {
	*httpserver.HTTPServer
	log  log.Log
	conn *mongo.Mongo
	coll *mongo.Collection
	c    *testutils.TestConfig
}

type body struct {
	UUID string `json:"UUID"`
}

func (s *server) Post(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	payload := body{}
	json.Unmarshal(data, &payload)
	s.coll.InsertOne(r.Context(), sample.GetRandomData(payload.UUID))
	w.WriteHeader(http.StatusCreated)
}

func (s *server) Get(c *gin.Context) {
	param := c.Query("UUID")
	data := sample.PIITestVal{}
	cur := s.coll.FindOne(c.Request.Context(), map[string]string{"UUID": param})
	cur.Decode(&data)
	w := c.Writer
	s.WriteJSON(c.Request.Context(), w, data)
}

func (s *server) Name(ctx context.Context) string {
	return "CSFLE"
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.conn.Shutdown(ctx)
}

func NewServer(t instrumentation.Tracer) *server {
	testutils.SetAWSConfig(t)
	ctx := GetCorrelationContext()
	loc := utils.GetEnv("SCHEME_LOCATION", "./db/mongo/csfle/sample/piischeme.json")
	file, err := os.Open(loc)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error opening scheme file", err, nil)
	}
	schemeByte, err := io.ReadAll(file)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error reading scheme file", err, nil)
	}
	scheme := string(schemeByte)
	kmsArn := ServerTestConfig.AWS.KMS
	dbName, collName := "GOTEST", "PII"
	kmsProvider, err := sample.GetKMSProvider(ctx, ServerTestLogger, kmsArn)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating kms", err, nil)
	}
	config := ServerTestConfig.CSFLE
	client, err := mongo.NewWithDefaultOptions(ServerTestLogger, t)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo client", err, nil)
	}
	dbScheme, err := csfle.SetEncryptionKey(ctx, ServerTestLogger, &scheme, client, config.KeyVaultNamespace, kmsProvider)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating scheme", err, nil)
	}
	config.KMSCredentials = kmsProvider.Credentials()
	config.SchemaMap = dbScheme
	conn, err := csfle.New(ServerTestLogger, t, *ServerTestConfig.CSFLE)
	if err != nil {
		ServerTestLogger.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		HTTPServer: httpserver.New(httpserver.WithTracer(t)), log: ServerTestLogger,
		conn: conn,
		coll: conn.Database(dbName).Collection(collName),
		c:    ServerTestConfig,
	}
	srv.RegisterOnShutdownHook(srv)
	r := srv.GetRouter().Group("/vault/v1")
	r.GET("", srv.Get)
	r.POST("", gin.WrapF(srv.Post))
	return srv
}
