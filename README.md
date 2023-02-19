<div align="center">

![Trip](https://github.com/philippta/trip/blob/assets/trip.png?raw=true)

[![Go Reference](https://pkg.go.dev/badge/github.com/philippta/trip.svg)](https://pkg.go.dev/github.com/philippta/trip) [![Go Report Card](https://goreportcard.com/badge/github.com/philippta/trip)](https://goreportcard.com/report/github.com/philippta/trip) [![MIT](https://img.shields.io/github/license/philippta/trip?style=flat)](https://github.com/philippta/trip/blob/master/LICENSE) [![Release](https://img.shields.io/github/release/philippta/trip.svg)](https://github.com/philippta/trip/releases)

Elegant middleware functions for your HTTP clients.
</div>

---

Trip is the enhancement for your HTTP clients:
- Authorize HTTP requests **once** and for all.
- Make requests **more resilient** against temporary failures.
- **Removes clutter** from your HTTP calls.
- Plugs easily into your existing HTTP clients.
- Zero dependencies.
- Tiny and readable codebase.

---

## Concepts

Trip is aimed to be used with the HTTP client and act as a middleware before any request goes out.

In a nutshell, the `http.Client` builds the HTTP request with its headers and body and hands it over to the transport. The transport then sends it off to the server and waits for the response. Once it got the response, it gives it back to the `http.Client`.

Trip intercepts the hand-over part between client and transport and can modify the request before it goes out. It can also inspect the response and take action like retrying a request on network errors.

## Installation

Trip requires [Go 1.18](https://go.dev/dl/) or higher. Use `go get` to install the library.

```
go get -u github.com/philippta/trip@latest
```

Next, import it into your application:
```go
import "github.com/philippta/trip"
```

## Usage

Trip can be easily used with any `http.Client` by creating a new instance and setting it as the `Transport` of the client.

```go
package main

import (
    "net/http"

    "github.com/philippta/trip"
)

func main() {
    var (
        attempts = 3
        delay    = 50 * time.Millisecond
    )

    client := &http.Client{
        Transport: trip.Default(
            // Auth
            trip.BearerToken("api-token"),
            trip.BasicAuth("username", "password"),

            // Headers
            trip.Header("Cache-Control", "no-cache")
            trip.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; ...")

            // Logging
            trip.Logger(log.Printf)
            trip.Logger(logrus.Infof)                  // github.com/sirupsen/logrus
            trip.Logger(zap.S().Infof)                 // github.com/uber-go/zap
            trip.Logger(zerolog.New(os.Stdout).Printf) // github.com/rs/zerolog

            // Retry
            trip.Retry(attempts, delay),
            trip.Retry(attempts, delay, http.StatusTooManyRequests),
            trip.Retry(attempts, delay, trip.RetryableStatusCodes...),

            // Idempotency
            trip.IdempotencyKey()
        ),
    }

    client.Get("http://example.com/")
}
```

## Examples

Listed below are some examples how to use Trip for various situations.

#### Authentication (Bearer Token, Basic Auth)

```go
func main() {
    t := trip.Default(
        trip.BearerToken("api-token"),
        trip.BasicAuth("username", "password"),
    )

    client := &http.Client{Transport: t}
    client.Get("http://example.com/")
}
```

#### Static HTTP Headers (User Agent, Cache Control, etc.)

```go
func main() {
    t := trip.Default(
        trip.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; ..."),
        trip.Header("Cache-Control", "max-age=86400"),
    )

    client := &http.Client{Transport: t}
    client.Get("http://example.com/")
}
```

#### Retries for flaky networks or API servers

```go
func main() {
    var (
        attempts = 3
        delay    = 150 * time.Millisecond
    )

    t := trip.Default(
        // Retries connection failures
        trip.Retry(attempts, retryDelay),

        // Retries connection failures and status codes
        trip.Retry(attempts, retryDelay, http.StatusTooManyRequests),

        // Retries connection failures and common retryable status codes
        trip.Retry(attempts, retryDelay, trip.RetryableStatusCodes...),
    )

    client := &http.Client{Transport: t}
    client.Get("http://example.com/")
}
```

#### Retrying with Idempotency Keys

```go
func main() {
    var (
        attempts = 3
        delay    = 150 * time.Millisecond
    )

    t := trip.Default(
        // Retries connection failures
        trip.Retry(attempts, retryDelay),

        // Generates idempotency keys for POST and PATCH requests
        trip.IdempotencyKey(),
    )

    client := &http.Client{Transport: t}
    client.Get("http://example.com/")
}
```

#### Logging with different libraries (log, zap, logrus, zerolog)

```go
func main() {
    t := trip.Default(
        trip.Logger(log.Printf),
        trip.Logger(logrus.Infof),                  // github.com/sirupsen/logrus
        trip.Logger(zap.S().Infof),                 // github.com/uber-go/zap
        trip.Logger(zerolog.New(os.Stdout).Printf), // github.com/rs/zerolog
    )

    client := &http.Client{Transport: t}
    client.Get("http://example.com/")

    // Example message:
    // POST http://example.com/ - 200 OK - 12.34ms
}
```

#### Custom Interceptors

```go
func main() {
    t := trip.Default(
        func (next http.RoundTripper) http.RoundTripper {
            return trip.RoundTripFunc(func(r *http.Request) (http.Response, error) {
                // before request
                resp, err := next.RoundTrip(r)
                // after request
                return resp, err
            })
        }
    )

    client := &http.Client{Transport: t}
    client.Get("http://example.com/")
}
```

#### Extending the default HTTP client

```go
func main() {
    t := trip.Default(
        trip.BearerToken("api-token"),
    )

    http.DefaultClient.Transport = t
    http.Get("http://example.com/")
}
```


## Retryable HTTP Status Codes

`trip.RetryableStatusCodes` holds a list of common HTTP status codes that are considered temporary and can be retried.

```go
trip.RetryableStatusCodes = []int{
	http.StatusRequestTimeout,      // 408
	http.StatusTooEarly,            // 425
	http.StatusTooManyRequests,     // 429
	http.StatusInternalServerError, // 500
	http.StatusBadGateway,          // 502
	http.StatusServiceUnavailable,  // 503
	http.StatusGatewayTimeout,      // 504
}
```

## Contributing

If you like to contribute to Trip by adding new features, improving the documentation or fixing bugs, feel free to open a [new issue](https://github.com/philippta/trip/issues).
