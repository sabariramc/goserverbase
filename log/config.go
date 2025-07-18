package log

import (
	"github.com/sabariramc/goserverbase/v6/env"
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

// GetDefaultConfig returns the new Config with values from environment variables or default values.
/*
	Environment Variables
	- SERVICE_NAME: Sets [ServiceName]
	- LOG__LEVEL: Sets [LogLevel], following are the valid options
		- TRACE
		- DEBUG
		- INFO
		- NOTICE
		- WARNING
		- ERROR
		- CRITICAL
		- EMERGENCY
	- LOG__FILE_TRACE: Sets [FileTrace]
	- LOG__WRITER: Sets the log writer for [Mux], supports following values by default, can be extended
		- CONSOLE
		- JSONL

For custom [LOG__WRITER] use [logwriter.AddLogWriter] before the package initialization
*/
func GetDefaultConfig() Config {
	writer := utils.GetEnv(env.LogWriter, "CONSOLE")
	w := logwriter.GetLogWriter(writer)
	if w == nil {
		w = logwriter.NewConsoleWriter()
	}
	return Config{
		ServiceName: utils.GetEnv(env.ServiceName, "default"),
		ModuleName:  "log",
		LogLevel:    message.GetLogLevelWithName(utils.GetEnv(env.LogLevel, "ERROR")),
		FileTrace:   utils.GetEnvBool(env.LogFileTrace, false),
		Mux:         NewDefaultLogMux(w),
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
