package retryhttp

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sabariramc/goserverbase/v6/log"
)

// Config contains the configuration settings for the HTTP client, including logging,
// retry policies, backoff strategies, and the HTTP client itself.
type Config struct {
	Log          log.Log       // Log is the logger used for logging HTTP client activities.
	RetryMax     uint          // RetryMax is the maximum number of retry attempts for failed requests.
	MinRetryWait time.Duration // MinRetryWait is the minimum duration to wait before retrying a failed request.
	MaxRetryWait time.Duration // MaxRetryWait is the maximum duration to wait before retrying a failed request.
	CheckRetry   CheckRetry    // CheckRetry is the function to determine if a request should be retried.
	Backoff      Backoff       // Backoff is the function to determine the wait duration between retries.
	Tracer       Tracer        // Tracer is used for tracing HTTP requests (assuming it's defined elsewhere).
	Client       *http.Client  // Client is the underlying HTTP client used to make requests.
}

// NewHTTPClient creates and configures a new HTTP client with custom transport settings.
func NewHTTPClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100        // MaxIdleConns sets the maximum number of idle connections across all hosts.
	t.MaxConnsPerHost = 100     // MaxConnsPerHost sets the maximum number of connections per host.
	t.MaxIdleConnsPerHost = 100 // MaxIdleConnsPerHost sets the maximum number of idle connections per host.
	return &http.Client{Transport: t}
}

// GetDefaultConfig returns a Config instance with default settings for the HTTP client.
func GetDefaultConfig() Config {
	return Config{
		Log:          log.New().NewResourceLogger("HTTPClient"), // Creates a new logger instance for the HTTP client.
		RetryMax:     4,                                         // Sets the maximum number of retry attempts to 4.
		MinRetryWait: time.Millisecond * 10,                     // Sets the minimum retry wait duration to 10 milliseconds.
		MaxRetryWait: time.Second * 5,                           // Sets the maximum retry wait duration to 5 seconds.
		CheckRetry:   retryablehttp.DefaultRetryPolicy,          // Uses the default retry policy.
		Backoff:      retryablehttp.DefaultBackoff,              // Uses the default backoff strategy.
		Client:       NewHTTPClient(),                           // Uses a custom HTTP client with specific transport settings.
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
