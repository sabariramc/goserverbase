package httpserver_test

import (
	"github.com/gin-gonic/gin"
	"github.com/sabariramc/goserverbase/v6/app/server/httpserver"
)

func Example() {
	srv := httpserver.New()
	srv.StartServer()
}

func Example_routes() {
	srv := httpserver.New()
	r := srv.GetRouter()
	r.Group("/test").GET("", func(ctx *gin.Context) {
		l := srv.GetLogger()
		l.Info(ctx, "test route")
	})
	srv.StartServer()
}
