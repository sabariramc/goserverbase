package httputil

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v6/log"
)

type Config struct {
	Log          log.Log
	RetryMax     uint
	MinRetryWait time.Duration
	MaxRetryWait time.Duration
	CheckRetry   CheckRetry
	Backoff      Backoff
	Tracer       Tracer
	Client       *http.Client
}

func newHTTP() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	return &http.Client{Transport: t}
}

func getDefaultConfig() Config {
	return Config{
		Log:          log.New().NewResourceLogger("HTTPClient"),
		RetryMax:     4,
		MinRetryWait: time.Millisecond * 10,
		MaxRetryWait: time.Second * 5,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		Client:       newHTTP(),
	}
}

// Option represents an option function for configuring the config struct.
type Option func(*Config)

// WithLog sets the log instance for the config.
func WithLog(log log.Log) Option {
	return func(c *Config) {
		c.Log = log
	}
}

// WithRetryMax sets the maximum number of retries for the config.
func WithRetryMax(retryMax uint) Option {
	return func(c *Config) {
		c.RetryMax = retryMax
	}
}

// WithMinRetryWait sets the minimum retry wait duration for the config.
func WithMinRetryWait(minRetryWait time.Duration) Option {
	return func(c *Config) {
		c.MinRetryWait = minRetryWait
	}
}

// WithMaxRetryWait sets the maximum retry wait duration for the config.
func WithMaxRetryWait(maxRetryWait time.Duration) Option {
	return func(c *Config) {
		c.MaxRetryWait = maxRetryWait
	}
}

// WithCheckRetry sets the check retry function for the config.
func WithCheckRetry(checkRetry CheckRetry) Option {
	return func(c *Config) {
		c.CheckRetry = checkRetry
	}
}

// WithBackoff sets the backoff strategy for the config.
func WithBackoff(backoff Backoff) Option {
	return func(c *Config) {
		c.Backoff = backoff
	}
}

// WithTracer sets the tracer instance for the config.
func WithTracer(tr Tracer) Option {
	return func(c *Config) {
		c.Tracer = tr
	}
}

// WithHTTPClient sets the HTTP client for the config.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.Client = client
	}
}
