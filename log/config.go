package log

import "github.com/sabariramc/goserverbase/v6/utils"

type Config struct {
	ServiceName string
	ModuleName  string
	LogLevel    LogLevel
	Mux         Mux
	FileTrace   bool
	Audit       AuditLogWriter
}

var defaultConfig = Config{
	ServiceName: utils.GetEnv("SERVICE_NAME", "default"),
	ModuleName:  "log",
	LogLevel:    GetLogLevel(utils.GetEnv("LOG__LEVEL", "default")),
	FileTrace:   false,
	Mux:         NewDefaultLogMux(),
	Audit:       nil,
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
		logLevel, ok := logLevelInverseMap[logLevelName]
		if !ok {
			logLevel = logLevelInverseMap["INFO"]
		}
		c.LogLevel = *logLevel
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
