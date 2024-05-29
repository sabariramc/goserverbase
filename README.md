# GO SERVER BASE

Micro framework for HTTP server and kafka client based microservices


## HTTP Server

Based on [gin](github.com/gin-gonic/gin)

### Basic

```go
srv := httpserver.New()
srv.StartServer()
```

The above server will have the following preconfigured routes

- `GET /meta/health`
- `GET /meta/status`
- `GET /meta/docs/*any`
- `GET /meta/static/*filepath`
- `HEAD /meta/static/*filepath`

OpenAPI documentation is configured at `GET /meta/docs/index.html`

### Basic with custom routes

```go
srv := httpserver.New()
r := srv.GetRouter() // returns gin.Engine
r.Group("/test").GET("", func(ctx *gin.Context) {
    l := srv.GetLogger()
    l.Info(ctx, "test route")
})
srv.StartServer()
```

### Custom Server

```go

/server/server.go

package server

import (
    ...
    "github.com/sabariramc/goserverbase/v6/app/server/httpserver"
    "github.com/sabariramc/goserverbase/v6/log"
    "github.com/sabariramc/goserverbase/v6/db/mongo"
    "github.com/sabariramc/goserverbase/v6/instrumentation"
)

type server struct {
	*httpserver.HTTPServer
	log        log.Log
	conn       *mongo.Mongo
}

func GetCorrelationContext() context.Context {
	ctx := context.WithValue(context.Background(), correlation.ContextKeyCorrelation, correlation.NewCorrelationParam(ServiceName))
	return ctx
}

func (s *server) test(c *gin.Context) {
    db := s.conn.Database("test")
    ...
}


func New(t instrumentation.Tracer) *server {
	ctx := GetCorrelationContext()
	conn, err := mongo.NewWithDefaultOptions(ServerTestLogger, t)
    log := log.New()
	if err != nil {
		log.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		HTTPServer: httpserver.New(httpserver.WithTracer(t)),
        log: log,
		conn: conn,
	}
	srv.RegisterHooks(conn)
	service := srv.GetRouter().Group("/service")
	echo := service.Group("/test")
	echo.POST("", srv.test)
    return srv
}

/cmd/main.go

package main

import server "..."


func main() {
	s := server.New(nil)
	s.StartServer()
}
```

#### With Datadog

```go
/cmd/main.go

package main

import (
	"log"

	server "..."
	"github.com/sabariramc/goserverbase/v6/instrumentation/contrib/ddtrace"
)

func main() {
	tr, err := ddtrace.Init()
	if err != nil {
		log.Fatal("tracer failed", err)
	}
	defer ddtrace.ShutDown()
	s := server.New(tr)
	s.StartServer()
}
```

## Kafka Client

Based on Based on [segmentio/kafka-go](github.com/segmentio/kafka-go)

### Basic

```go
srv := kafkaclient.New()
srv.AddHandler(context.Background(), "gobase.test.topic1", func(ctx context.Context, m *kafka.Message) error {
    return nil
})
srv.AddHandler(context.Background(), "gobase.test.topic2", func(ctx context.Context, m *kafka.Message) error {
    return &errors.CustomError{ErrorCode: "gobase.test.error", ErrorMessage: "error sample"}
})
srv.StartConsumer()
```

This client subscribes to the following topics `gobase.test.topic1` and `gobase.test.topic2`, passes the messages to the handler function

### Custom

```go
/server/server.go

package server

import (
    ...
    "github.com/sabariramc/goserverbase/v6/app/server/kafkaclient"
    "github.com/sabariramc/goserverbase/v6/log"
    "github.com/sabariramc/goserverbase/v6/db/mongo"
    "github.com/sabariramc/goserverbase/v6/instrumentation"
)

type server struct {
	*kafkaclient.KafkaClient
	log        log.Log
	conn       *mongo.Mongo
}

func (s *server) test(ctx context.Context, event *kafka.Message) error {
	data := make(map[string]any)
	err := event.LoadBody(&data)
	if err != nil {
		return fmt.Errorf("server.test: error loading body: %w", err)
	}
    coll := s.conn.Database("GOBaseTest").Collection("TestColl")
	coll.InsertOne(ctx, data)
	...
	return nil
}


func New(t instrumentation.Tracer) *server {
	ctx := GetCorrelationContext()
	conn, err := mongo.NewWithDefaultOptions(ServerTestLogger, t)
    log := log.New()
	if err != nil {
		log.Emergency(ctx, "error creating mongo connection", err, nil)
	}
	srv := &server{
		KafkaClient: kafkaclient.New(kafkaclient.WithTracer(t)),
		conn:        conn,
        log: log,
	}
	srv.RegisterHooks(conn)
	srv.AddHandler(GetCorrelationContext(), "gobase.test.topic", srv.test)
	return srv
}

/cmd/main.go

package main

import (
	server "..."
)

func main() {
	s := server.NewServer(nil)
	s.StartConsumer()
}
```

#### With Datadog

```go
package main

import (
	server "..."
	"github.com/sabariramc/goserverbase/v6/instrumentation/contrib/ddtrace"
)

func main() {
	tr, err := ddtrace.Init()
	if err != nil {
		log.Fatal("tracer failed", err)
	}
	defer ddtrace.ShutDown()
	s := server.NewServer(tr)
	s.StartConsumer()
}
```

For complete example implementation are under folder `app/server/httpserver/test` and `app/server/kafkaclient/test`

