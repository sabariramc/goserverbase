package main

import (
	"context"
	"fmt"
	stlLog "log"
	"net/http"

	"github.com/sabariramc/goserverbase/v2/baseapp/test/server"
)

func main() {
	s := server.NewServer()
	s.GetLogger().Notice(context.TODO(), fmt.Sprintf("Server starting at %v", s.GetPort()), nil)
	stlLog.Fatal(http.ListenAndServe(s.GetPort(), s))
}
