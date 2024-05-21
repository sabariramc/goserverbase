package log

import (
	"github.com/sabariramc/goserverbase/v6/envvariables"
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/log/message"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// Config represents the configuration options for the logger.
type Config struct {
	ServiceName string           // ServiceName represents the name of the service.
	ModuleName  string           // ModuleName represents the name of the module.
	LogLevel    message.LogLevel // LogLevel represents the log level.
	Mux         Mux              // Mux represents the multiplexer for handling log messages.
	FileTrace   bool             // FileTrace indicates whether file tracing is enabled.
	Audit       AuditLogWriter   // Audit represents the audit log writer.
}

// getDefaultConfig returns the default configuration for the logger.
func getDefaultConfig() Config {
	return Config{
		ServiceName: utils.GetEnv(envvariables.ServiceName, "default"),
		ModuleName:  "log",
		LogLevel:    message.GetLogLevelWithName(utils.GetEnv(envvariables.LogLevel, "ERROR")),
		FileTrace:   false,
		Mux:         NewDefaultLogMux(logwriter.NewConsoleWriter()),
		Audit:       nil,
	}
}

// Option represents an option function for configuring the logger.
type Option func(*Config)

// WithServiceName sets the service name for the logger.
func WithServiceName(serviceName string) Option {
	return func(c *Config) {
		c.ServiceName = serviceName
	}
}

// WithModuleName sets the module name for the logger.
func WithModuleName(moduleName string) Option {
	return func(c *Config) {
		c.ModuleName = moduleName
	}
}

// WithLogLevelName sets the log level name for the logger.
func WithLogLevelName(logLevelName string) Option {
	return func(c *Config) {
		c.LogLevel = message.GetLogLevelWithName(logLevelName)
	}
}

// WithMux sets the Mux for the logger.
func WithMux(mux Mux) Option {
	return func(c *Config) {
		c.Mux = mux
	}
}

// WithFileTrace sets the file trace flag for the logger.
func WithFileTrace(fileTrace bool) Option {
	return func(c *Config) {
		c.FileTrace = fileTrace
	}
}

// WithAudit sets the audit log writer for the logger.
func WithAudit(audit AuditLogWriter) Option {
	return func(c *Config) {
		c.Audit = audit
	}
}
