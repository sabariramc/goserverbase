package httpserver

import baseapp "github.com/sabariramc/goserverbase/v4/app"

type LogConfig struct {
	AuthHeaderKeyList []string
	ContentLength     int64
}

type DocumentationConfig struct {
	DocHost           string
	SwaggerRootFolder string
}

type HTTP2Config struct {
	PublicKeyPath  string
	PrivateKeyPath string
}

type HTTPServerConfig struct {
	baseapp.ServerConfig
	DocumentationConfig
	*HTTP2Config
	Host string
	Port string
	Log  *LogConfig
}
