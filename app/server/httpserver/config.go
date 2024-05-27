// Package httpserver provides configuration options for an HTTP server.
package httpserver

import (
	baseapp "github.com/sabariramc/goserverbase/v6/app"
	"github.com/sabariramc/goserverbase/v6/env"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// MaskConfig holds the configuration for masking headers in log messages.
type MaskConfig struct {
	HeaderKeyList []string // List of header keys to mask before logging request
}

// GetDefaultMaskConfig returns the default MaskConfig with values from environment variables or default values.
/*
	Environment Variables
	- HTTP_SERVER__MASK__HEADER_KEY_LIST: Sets [HeaderKeyList]
*/
func GetDefaultMaskConfig() *MaskConfig {
	return &MaskConfig{
		HeaderKeyList: utils.GetEnvAsSlice(env.HTTPServerMaskHeaderKeyList, []string{"Authorization", "x-api-key"}, ","),
	}
}

// DocumentationConfig holds the configuration for serving documentation.
type DocumentationConfig struct {
	DocHost    string // Host for the documentation server
	RootFolder string // Local disk folder for the OpenAPI documentation server
}

// GetDocumentationConfig returns the default DocumentationConfig with values from environment variables or default values.
/*
	Environment Variables
	- HTTP_SERVER__DOC_HOST: Sets [DocHost]
	- HTTP_SERVER__DOC_ROOT_FOLDER: Sets [RootFolder]
*/
func GetDocumentationConfig() *DocumentationConfig {
	return &DocumentationConfig{
		DocHost:    utils.GetEnv(env.HTTPServerDocHost, "localhost"),
		RootFolder: utils.GetEnv(env.HTTPServerDocRootFolder, "/docs"),
	}
}

// TLSConfig holds the configuration for HTTPS.
type TLSConfig struct {
	PublicKeyPath  string // Local disk path for the public key
	PrivateKeyPath string // Local disk path for the private key
}

// GetDefaultTLSConfig returns the default TLSConfig with values from environment variables or default values.
/*
	Environment Variables
	- HTTP_SERVER__TLS_PUBLIC_KEY: Sets [PublicKeyPath]
	- HTTP_SERVER__TLS_PRIVATE_KEY: Sets [PrivateKeyPath]
*/
func GetDefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		PublicKeyPath:  utils.GetEnv(env.HTTPServerTLSPublicKey, "publickey.cer"),
		PrivateKeyPath: utils.GetEnv(env.HTTPServerTLSPrivateKey, "privatekey.cer"),
	}
}

// Config holds the configuration for the HTTP server.
type Config struct {
	*baseapp.Config
	*DocumentationConfig
	*TLSConfig
	Host   string      // Host address
	Port   string      // Port number
	Log    log.Log     // Logger instance
	Mask   *MaskConfig // Configuration for masking headers
	Tracer Tracer      // Tracer instance
}

// GetDefaultConfig returns the default HTTPServerConfig with values from environment variables or default values.
/*
	Environment Variables
	- HTTP_SERVER__HOST: Sets [Host]
	- HTTP_SERVER__PORT: Sets [Port]
*/
func GetDefaultConfig() *Config {
	return &Config{
		Config:              baseapp.GetDefaultConfig(),
		DocumentationConfig: GetDocumentationConfig(),
		TLSConfig:           GetDefaultTLSConfig(),
		Mask:                GetDefaultMaskConfig(),
		Log:                 log.New(log.WithModuleName("HTTPServer")),
		Host:                utils.GetEnv(env.HTTPServerHost, "0.0.0.0"),
		Port:                utils.GetEnv(env.HTTPServerPort, "3000"),
	}
}

// Option represents a function that applies a configuration option to HTTPServerConfig.
type Option func(*Config)

// WithBaseAppConfig sets the baseapp.Config embedded field of HTTPServerConfig.
func WithBaseAppConfig(baseCfg *baseapp.Config) Option {
	return func(c *Config) {
		c.Config = baseCfg
	}
}

// WithDocumentationConfig sets the DocumentationConfig embedded field of HTTPServerConfig.
func WithDocumentationConfig(docCfg *DocumentationConfig) Option {
	return func(c *Config) {
		c.DocumentationConfig = docCfg
	}
}

// WithHTTP2Config sets the TLSConfig embedded field of HTTPServerConfig.
func WithHTTP2Config(http2Cfg *TLSConfig) Option {
	return func(c *Config) {
		c.TLSConfig = http2Cfg
	}
}

// WithHost sets the Host field of HTTPServerConfig.
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the Port field of HTTPServerConfig.
func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithLog sets the Log field of HTTPServerConfig.
func WithLog(log log.Log) Option {
	return func(c *Config) {
		c.Log = log
	}
}

// WithMask sets the Log field of HTTPServerConfig.
func WithMask(m MaskConfig) Option {
	return func(c *Config) {
		c.Mask = &m
	}
}

// WithTracer sets the Tracer field of HTTPServerConfig.
func WithTracer(t Tracer) Option {
	return func(c *Config) {
		c.Tracer = t
	}
}
