<div align="center">

![Trip](https://github.com/philippta/trip/blob/assets/trip.png?raw=true)

[![Go Reference](https://pkg.go.dev/badge/github.com/philippta/trip.svg)](https://pkg.go.dev/github.com/philippta/trip) [![Go Report Card](https://goreportcard.com/badge/github.com/philippta/trip)](https://goreportcard.com/report/github.com/philippta/trip) [![MIT](https://img.shields.io/github/license/philippta/trip)](https://img.shields.io/github/license/philippta/trip) ![Code size](https://img.shields.io/github/languages/code-size/philippta/trip)

Useful round trip functions for your HTTP clients.
</div>

---

### Installation

```
$ go get github.com/philippta/trip@latest
```

### Basic usage

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
        RoundTripper: trip.Default(
            // Auth
            trip.BearerToken("eyJhbGc..."),
            trip.BasicAuth("username", "password"),

            // Retry
            trip.Retry(attempts, delay),
            trip.RetryNetwork(attempts, delay),
            trip.RetryStatusCodes(attempts, delay,
                http.StatusServiceUnavailable,
                http.TooManyRequests
            ),
        ),
    }
}
```

### Examples

```go
// Create a transport for your http.Client
trip.Default(...)
trip.New(http.DefaultTransport, ...) // same as trip.Default(...)

// Set "Authorization" header on every request
trip.BearerToken("token")          // Authorization: Bearer token
trip.BasicAuth("user", "password") // Authorization: Basic dXNlcjpwYXNzd29yZA==

// Retry for problematic networks and sensible servers
trip.Retry(attempts, delay)                      // retries on network errors and retryable HTTP status codes
trip.RetryNetwork(attempts, delay)               // retries on network errors
trip.RetryStatusCodes(attempts, delay, 429, ...) // retires on network errors and given HTTP status codes

// Create a HTTP client with bearer token auth
// and sane retry behaviour based on http.DefaultTransport
client := &http.Client{
    Transport: trip.Default(
        trip.BearerToken("token"),
        trip.Retry(3, 50 * time.Millisecond),
    )
}
```

### License

MIT
