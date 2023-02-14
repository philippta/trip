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
		token    = "abc123"
		expected = "Bearer abc123"
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		assertEqual(t, r.Header.Get("Authorization"), expected)
		return nil, nil
	}, trip.BearerToken(token))
}

func TestBasicAuth(t *testing.T) {
	var (
		username = "username"
		password = "password"
		expected = "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		assertEqual(t, r.Header.Get("Authorization"), expected)
		return nil, nil
	}, trip.BasicAuth(username, password))
}

func TestRetryNetwork(t *testing.T) {
	var (
		calls []time.Time

		attempts = 3
		delay    = 2 * time.Millisecond
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, time.Now())
		return nil, errors.New("network error")
	}, trip.Retry(attempts, delay))

	assertEqual(t, len(calls), attempts)
	assertTimeRange(t, calls[0], calls[1], delay, time.Millisecond)
}

func TestRetryStatusCodes(t *testing.T) {
	var (
		calls []time.Time

		attempts = 3
		delay    = 2 * time.Millisecond
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, time.Now())
		return &http.Response{StatusCode: http.StatusBadGateway}, nil
	}, trip.Retry(attempts, delay, trip.RetryableStatusCodes...))

	assertEqual(t, len(calls), attempts)
	assertTimeRange(t, calls[0], calls[1], delay, time.Millisecond)
}

func TestRetryStatusCodesSkipped(t *testing.T) {
	var (
		calls []time.Time

		attempts = 3
		delay    = 2 * time.Millisecond
		codes    = []int{http.StatusTooManyRequests}
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, time.Now())
		return &http.Response{StatusCode: http.StatusBadGateway}, nil
	}, trip.Retry(attempts, delay, codes...))

	assertEqual(t, len(calls), 1)
}

func roundTrip(f trip.RoundTripperFunc, trips ...trip.TripFunc) {
	req := httptest.NewRequest("GET", "/", nil)
	transport := trip.New(f, trips...)
	transport.RoundTrip(req)
}

func assertEqual[T comparable](t *testing.T, a T, b T) {
	if a != b {
		t.Errorf("got: %v, expected: %v", a, b)
	}
}

func assertTimeRange(t *testing.T, a time.Time, b time.Time, expectedDiff, timeRange time.Duration) {
	diff := b.Sub(a)
	if diff-expectedDiff > timeRange || diff-expectedDiff < -timeRange {
		t.Errorf("took: %v, expected: %v, not in range of %v", diff, expectedDiff, timeRange)
	}
}
