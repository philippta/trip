package trip_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestHeader(t *testing.T) {
	var (
		key   = "X-Foo"
		value = "bar"
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		assertEqual(t, r.Header.Get(key), value)
		return nil, nil
	}, trip.Header(key, value))
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

func TestUserAgent(t *testing.T) {
	var (
		userAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		assertEqual(t, r.Header.Get("User-Agent"), userAgent)
		return nil, nil
	}, trip.UserAgent(userAgent))
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
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(""))}, nil
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
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(""))}, nil
	}, trip.Retry(attempts, delay, codes...))

	assertEqual(t, len(calls), 1)
}

func TestIdempotencyKey(t *testing.T) {
	var (
		idems []string

		attempts = 3
		delay    = 2 * time.Millisecond
	)

	roundTrip(func(r *http.Request) (*http.Response, error) {
		idems = append(idems, r.Header.Get("Idempotency-Key"))
		return nil, errors.New("network error")
	}, trip.Retry(attempts, delay), trip.IdempotencyKey())

	assertEqual(t, len(idems), attempts)
	assertNotEqual(t, idems[0], "")
	assertEqual(t, idems[0], idems[1])
}

func TestLogger(t *testing.T) {
	logf := func(format string, v ...any) {
		msg := fmt.Sprintf(format, v...)
		assertPrefix(t, msg, "POST http://example.com/foo?bar=yes - 200 OK -")
	}

	roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{Status: "200 OK", StatusCode: 200, Body: io.NopCloser(strings.NewReader("foo"))}, nil
	}, trip.Logger(logf))
}

func TestLoggerError(t *testing.T) {
	logf := func(format string, v ...any) {
		msg := fmt.Sprintf(format, v...)
		assertPrefix(t, msg, `POST http://example.com/foo?bar=yes - error:"network error" -`)
	}

	roundTrip(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}, trip.Logger(logf))
}

func roundTrip(f trip.RoundTripperFunc, trips ...trip.TripFunc) {
	req := httptest.NewRequest("POST", "http://example.com/foo?bar=yes", nil)
	transport := trip.New(f, trips...)
	transport.RoundTrip(req)
}

func noop(r *http.Request) (*http.Response, error) {
	return nil, nil
}

func assertEqual[T comparable](t *testing.T, a T, b T) {
	if a != b {
		t.Errorf("got: %v, expected: %v", a, b)
	}
}

func assertPrefix(t *testing.T, a, b string) {
	if !strings.HasPrefix(a, b) {
		t.Errorf("got: %v, expected to start with: %v", a, b)
	}
}

func assertNotEqual[T comparable](t *testing.T, a T, b T) {
	if a == b {
		t.Errorf("got: %v and %v, expected something else", a, b)
	}
}

func assertTimeRange(t *testing.T, a time.Time, b time.Time, expectedDiff, timeRange time.Duration) {
	diff := b.Sub(a)
	if diff-expectedDiff > timeRange || diff-expectedDiff < -timeRange {
		t.Errorf("took: %v, expected: %v, not in range of %v", diff, expectedDiff, timeRange)
	}
}
