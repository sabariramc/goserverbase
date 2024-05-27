package baseapp

import (
	"github.com/sabariramc/goserverbase/v6/envvariables"
	"github.com/sabariramc/goserverbase/v6/log"
	"github.com/sabariramc/goserverbase/v6/notifier"
	"github.com/sabariramc/goserverbase/v6/utils"
)

// Config holds the configuration for the base app.
type Config struct {
	ServiceName string
	Log         log.Log
	Notifier    notifier.Notifier
}

// Option represents a function that applies a configuration option to Config.
type Option func(*Config)

// GetDefaultConfig creates a new default Config with values from environment variables or default values.
func GetDefaultConfig() *Config {
	return &Config{
		ServiceName: utils.GetEnv(envvariables.ServiceName, "default"),
		Log:         log.New(log.WithModuleName("BaseApp")),
	}
}

// WithServiceName sets the ServiceName field of Config.
func WithServiceName(serviceName string) Option {
	return func(c *Config) {
		c.ServiceName = serviceName
	}
}

// WithLog sets the Log field of Config.
func WithLog(log log.Log) Option {
	return func(c *Config) {
		c.Log = log
	}
}

// WithNotifier sets the Notifier field of Config.
func WithNotifier(notifier notifier.Notifier) Option {
	return func(c *Config) {
		c.Notifier = notifier
	}
}
