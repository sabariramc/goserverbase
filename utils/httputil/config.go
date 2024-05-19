package httputil

import (
	"net/http"
	"time"

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

var defaultOption = config{
    log: ,
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
