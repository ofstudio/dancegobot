// Package ratelimit - provides a rate-limited http client.
//
// Example:
//
//	rps := 25 // requests per second
//	timeout := 30 * time.Second
//	httpClient := ratelimit.Client(rps, timeout)
package ratelimit
