package ratelimit

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Client returns a rate-limited http client
func Client(rps int, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &transport{
			RoundTripper: http.DefaultTransport,
			limiter:      rate.NewLimiter(rate.Limit(rps), 1),
		},
		Timeout: timeout,
	}
}

// transport is a rate-limited http transport.
type transport struct {
	http.RoundTripper
	limiter *rate.Limiter
}

// RoundTrip implements http.RoundTripper with rate limiting.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := t.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	return t.RoundTripper.RoundTrip(req)
}
