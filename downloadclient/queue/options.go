package queue

import (
	"net/http"
	"time"
)

const defaultTimeout = 10 * time.Second

type options struct {
	timeout    time.Duration
	userAgent  string
	httpClient *http.Client
}

// Option configures options for Take.
type Option func(*options)

// WithTimeout bounds the queue observation operation. Default is 10s.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

// WithUserAgent sets the User-Agent header for HTTP requests.
func WithUserAgent(ua string) Option {
	return func(o *options) {
		o.userAgent = ua
	}
}

// WithHTTPClient sets a custom HTTP client for requests.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) {
		o.httpClient = c
	}
}
