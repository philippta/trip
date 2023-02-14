package trip

import (
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

// RetryableStatusCodes contains common HTTP status codes
// that are considered temporary and can be retried.
var RetryableStatusCodes = []int{
	http.StatusRequestTimeout,
	http.StatusTooEarly,
	http.StatusTooManyRequests,
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

// TripFunc is function for wrapping http.RoundTrippers.
type TripFunc func(http.RoundTripper) http.RoundTripper

// RoundTripperFunc implements http.RoundTripper for convenient usage.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip satisfies http.RoundTripper and calls fn.
func (fn RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// New creates a new http.RoundTripper by wrapping a given transport
// with the provided middleware/trip functions.
// If transport is nil, the http.DefaultTransport is used.
func New(transport http.RoundTripper, trips ...TripFunc) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	for _, trip := range trips {
		transport = trip(transport)
	}
	return transport
}

// Default creates a new http.RoundTripper based on http.DefaultTransport.
func Default(trips ...TripFunc) http.RoundTripper {
	return New(nil, trips...)
}

// BearerToken sets the `Authorization` header on every request to `Bearer <token>`.
func BearerToken(token string) TripFunc {
	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Set("Authorization", "Bearer "+token)
			return t.RoundTrip(r)
		})
	}
}

// BasicAuth sets the `Authorization` header on every request to `Basic <encoded-username-and-password>`.
func BasicAuth(username, password string) TripFunc {
	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Set("Authorization", "Basic "+encoded)
			return t.RoundTrip(r)
		})
	}
}

// Retry retries a failed HTTP request a given number of times and applies a fixed delay
// inbetween calls. Optionally a list of HTTP status codes can be provided that are
// considered as failure case.
// This can be used in combination with RetryableStatusCodes.
func Retry(attempts int, delay time.Duration, statusCodes ...int) TripFunc {
	retryable := func(statusCode int) bool {
		for _, code := range statusCodes {
			if statusCode == code {
				return true
			}
		}
		return false
	}

	drain := func(resp *http.Response) {
		if resp == nil || resp.Body == nil {
			return
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error

			for i := 0; i < attempts; i++ {
				resp, err = t.RoundTrip(r)
				if err == nil && !retryable(resp.StatusCode) {
					break
				}
				drain(resp)
				time.Sleep(delay)
			}

			return resp, err
		})
	}
}
