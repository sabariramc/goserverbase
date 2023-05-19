package httpserver

import baseapp "github.com/sabariramc/goserverbase/v2/app"

type LogConfig struct {
	AuthHeaderKeyList []string
	ContentLength     int64
}

type HttpServerConfig struct {
	*baseapp.ServerConfig
	Host string
	Port string
	Log  *LogConfig
}
