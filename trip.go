package trip

import (
	"encoding/base64"
	"net/http"
	"time"
)

type TripFunc func(http.RoundTripper) http.RoundTripper

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func New(transport http.RoundTripper, trips ...TripFunc) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	for _, trip := range trips {
		transport = trip(transport)
	}
	return transport
}

func Default(trips ...TripFunc) http.RoundTripper {
	return New(nil, trips...)
}

func BearerToken(token string) TripFunc {
	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Set("Authorization", "Bearer "+token)
			return t.RoundTrip(r)
		})
	}
}

func BasicAuth(username, password string) TripFunc {
	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Set("Authorization", "Basic "+encoded)
			return t.RoundTrip(r)
		})
	}
}

func Retry(attempts int, delay time.Duration) TripFunc {
	retryableStatusCodes := []int{408, 425, 429, 500, 502, 503, 504}
	return RetryStatusCodes(attempts, delay, retryableStatusCodes...)
}

func RetryNetwork(attempts int, delay time.Duration) TripFunc {
	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error

			for i := 0; i < attempts; i++ {
				resp, err = t.RoundTrip(r)
				if err == nil {
					break
				}
				time.Sleep(delay)
			}

			return resp, err
		})
	}
}

func RetryStatusCodes(attempts int, delay time.Duration, statusCodes ...int) TripFunc {
	retyable := func(statusCode int) bool {
		for _, code := range statusCodes {
			if statusCode == code {
				return true
			}
		}
		return false
	}

	return func(t http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			var resp *http.Response
			var err error

			for i := 0; i < attempts; i++ {
				resp, err = t.RoundTrip(r)
				if err == nil && !retyable(resp.StatusCode) {
					break
				}
				time.Sleep(delay)
			}

			return resp, err
		})
	}
}
