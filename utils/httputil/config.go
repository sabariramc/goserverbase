package httputil

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v6/log"
)

type config struct {
	log          log.Log
	retryMax     uint
	minRetryWait time.Duration
	maxRetryWait time.Duration
	checkRetry   CheckRetry
	backoff      Backoff
	tr           Tracer
	c            *http.Client
}

func newHTTP() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	return &http.Client{Transport: t}
}

var defaultConfig = config{
	log:          log.New().NewResourceLogger("HTTPClient"),
	retryMax:     4,
	minRetryWait: time.Millisecond * 10,
	maxRetryWait: time.Second * 5,
	checkRetry:   retryablehttp.DefaultRetryPolicy,
	backoff:      retryablehttp.DefaultBackoff,
	c:            newHTTP(),
}

// Option represents an option function for configuring the config struct.
type Option func(*config)

// WithLog sets the log instance for the config.
func WithLog(log log.Log) Option {
	return func(c *config) {
		c.log = log
	}
}

// WithRetryMax sets the maximum number of retries for the config.
func WithRetryMax(retryMax uint) Option {
	return func(c *config) {
		c.retryMax = retryMax
	}
}

// WithMinRetryWait sets the minimum retry wait duration for the config.
func WithMinRetryWait(minRetryWait time.Duration) Option {
	return func(c *config) {
		c.minRetryWait = minRetryWait
	}
}

// WithMaxRetryWait sets the maximum retry wait duration for the config.
func WithMaxRetryWait(maxRetryWait time.Duration) Option {
	return func(c *config) {
		c.maxRetryWait = maxRetryWait
	}
}

// WithCheckRetry sets the check retry function for the config.
func WithCheckRetry(checkRetry CheckRetry) Option {
	return func(c *config) {
		c.checkRetry = checkRetry
	}
}

// WithBackoff sets the backoff strategy for the config.
func WithBackoff(backoff Backoff) Option {
	return func(c *config) {
		c.backoff = backoff
	}
}

// WithTracer sets the tracer instance for the config.
func WithTracer(tr Tracer) Option {
	return func(c *config) {
		c.tr = tr
	}
}

// WithHTTPClient sets the HTTP client for the config.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.c = client
	}
}
