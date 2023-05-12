package httpserver

import baseapp "github.com/sabariramc/goserverbase/v2/app"

type HttpServerConfig struct {
	*baseapp.ServerConfig
	Host              string
	Port              string
	AuthHeaderKeyList []string
}
