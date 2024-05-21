package log

import (
	"github.com/sabariramc/goserverbase/v6/log/logwriter"
	"github.com/sabariramc/goserverbase/v6/log/message"
	"github.com/sabariramc/goserverbase/v6/utils"
)

type Config struct {
	ServiceName string
	ModuleName  string
	LogLevel    message.LogLevel
	Mux         Mux
	FileTrace   bool
	Audit       AuditLogWriter
}

func getDefaultConfig() Config {
	return Config{
		ServiceName: utils.GetEnv("SERVICE_NAME", "default"),
		ModuleName:  "log",
		LogLevel:    message.GetLogLevelWithName(utils.GetEnv("LOG__LEVEL", "ERROR")),
		FileTrace:   false,
		Mux:         NewDefaultLogMux(logwriter.NewConsoleWriter()),
		Audit:       nil,
	}
}

// Option represents an option function for configuring the Logger struct.
type Option func(*Config)

// WithServiceName sets the service name for Logger.
func WithServiceName(serviceName string) Option {
	return func(c *Config) {
		c.ServiceName = serviceName
	}
}

// WithModuleName sets the module name for Logger.
func WithModuleName(moduleName string) Option {
	return func(c *Config) {
		c.ModuleName = moduleName
	}
}

// WithLogLevelName sets the log level name for Logger.
func WithLogLevelName(logLevelName string) Option {
	return func(c *Config) {
		c.LogLevel = message.GetLogLevelWithName(logLevelName)
	}
}

// WithMux sets the Mux for Logger.
func WithMux(mux Mux) Option {
	return func(c *Config) {
		c.Mux = mux
	}
}

// WithFileTrace sets the file trace flag for Logger.
func WithFileTrace(fileTrace bool) Option {
	return func(c *Config) {
		c.FileTrace = fileTrace
	}
}

// WithAudit sets the Audit log writer for Logger.
func WithAudit(audit AuditLogWriter) Option {
	return func(c *Config) {
		c.Audit = audit
	}
}
