package trip_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/philippta/trip"
)

func ExampleDefault() {
	var (
		attempts = 3
		delay    = 50 * time.Millisecond
		apiToken = os.Getenv("API_TOKEN")
	)

	client := &http.Client{
		Transport: trip.Default(
			trip.BearerToken(apiToken),
			trip.Retry(attempts, delay, trip.RetryableStatusCodes...),
		),
	}

	client.Get("https://api.example.com/endpoint")
}

func TestBearerToken(t *testing.T) {
	var (
		token  = "abc123"
		expect = "Bearer abc123"
		check  = func(r *http.Request) {
			if r.Header.Get("Authorization") != expect {
				t.Errorf("auth header, expected: %q, got: %q", expect, r.Header.Get("Authorization"))
			}
		}
	)

	trans := trip.New(requestTripper(check), trip.BearerToken(token))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))
}

func TestBasicAuth(t *testing.T) {
	var (
		username = "username"
		password = "password"
		expect   = "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
		check    = func(r *http.Request) {
			if r.Header.Get("Authorization") != expect {
				t.Errorf("auth header, expected: %q, got: %q", expect, r.Header.Get("Authorization"))
			}
		}
	)

	trans := trip.New(requestTripper(check), trip.BasicAuth(username, password))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))
}

func TestRetryNetwork(t *testing.T) {
	var (
		expectedAttempts = 3
		recordedAttempts = 0
		resp             = func() (*http.Response, error) {
			recordedAttempts++
			return nil, errors.New("network error")
		}
	)

	trans := trip.New(responseTripper(resp), trip.Retry(expectedAttempts, 0))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))

	if expectedAttempts != recordedAttempts {
		t.Errorf("retry attempts, expected: %d, got: %d", expectedAttempts, recordedAttempts)
	}
}

func TestRetryDelay(t *testing.T) {
	var (
		firstCall        time.Time
		secondCall       time.Time
		expectedAttempts = 2
		recordedAttempts = 0
		resp             = func() (*http.Response, error) {
			recordedAttempts++
			if recordedAttempts == 1 {
				firstCall = time.Now()
			} else {
				secondCall = time.Now()
			}
			return nil, errors.New("network error")
		}
	)

	trans := trip.New(responseTripper(resp), trip.Retry(expectedAttempts, 5*time.Millisecond))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))

	if expectedAttempts != recordedAttempts {
		t.Errorf("retry attempts, expected: %d, got: %d", expectedAttempts, recordedAttempts)
	}

	delayDiff := secondCall.Sub(firstCall)
	if delayDiff < 5*time.Millisecond {
		t.Errorf("retry delay, expected greater 5m, got: %v", delayDiff)
	}
}

func TestRetryStatusCodes(t *testing.T) {
	var (
		expectedAttempts = 3
		recordedAttempts = 0
		resp             = func() (*http.Response, error) {
			recordedAttempts++
			return &http.Response{StatusCode: http.StatusBadGateway}, nil
		}
	)

	trans := trip.New(responseTripper(resp), trip.Retry(expectedAttempts, 0, http.StatusBadGateway))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))

	if expectedAttempts != recordedAttempts {
		t.Errorf("retry attempts, expected: %d, got: %d", expectedAttempts, recordedAttempts)
	}
}

func TestRetryStatusCodesSkipped(t *testing.T) {
	var (
		expectedAttempts = 1
		recordedAttempts = 0
		resp             = func() (*http.Response, error) {
			recordedAttempts++
			return &http.Response{StatusCode: http.StatusBadGateway}, nil
		}
	)

	trans := trip.New(responseTripper(resp), trip.Retry(expectedAttempts, 0, http.StatusTooManyRequests))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))

	if expectedAttempts != recordedAttempts {
		t.Errorf("retry attempts, expected: %d, got: %d", expectedAttempts, recordedAttempts)
	}
}

func TestRetryNetworkAndStatusCode(t *testing.T) {
	var (
		expectedAttempts = 3
		recordedAttempts = 0
		resp             = func() (*http.Response, error) {
			recordedAttempts++
			if recordedAttempts == 1 {
				return nil, errors.New("network error")
			}
			return &http.Response{StatusCode: http.StatusTooManyRequests}, nil
		}
	)

	trans := trip.New(responseTripper(resp), trip.Retry(expectedAttempts, 0, trip.RetryableStatusCodes...))
	trans.RoundTrip(httptest.NewRequest("GET", "/", nil))

	if expectedAttempts != recordedAttempts {
		t.Errorf("retry attempts, expected: %d, got: %d", expectedAttempts, recordedAttempts)
	}
}

func requestTripper(fn func(*http.Request)) http.RoundTripper {
	return trip.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		fn(r)
		return nil, nil
	})
}

func responseTripper(fn func() (*http.Response, error)) http.RoundTripper {
	return trip.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return fn()
	})
}
