<div align="center">

![Trip](https://github.com/philippta/trip/blob/assets/trip.png?raw=true)

[![Go Reference](https://pkg.go.dev/badge/github.com/philippta/trip.svg)](https://pkg.go.dev/github.com/philippta/trip) [![Go Report Card](https://goreportcard.com/badge/github.com/philippta/trip)](https://goreportcard.com/report/github.com/philippta/trip) [![MIT](https://img.shields.io/github/license/philippta/trip?style=flat)](https://github.com/philippta/trip/blob/master/LICENSE) [![Release](https://img.shields.io/github/release/philippta/trip.svg)](https://github.com/philippta/trip/releases)

Elegant middleware functions for your HTTP clients.
</div>

---

Trip enhances your HTTP clients in very elegant ways:
- Authorize HTTP requests **once** and for all.
- Make requests **more resilient** against temporary failures.
- **Removes clutter** from your HTTP calls.
- Plugs easily into your exisiting HTTP clients.
- Zero dependencies.
- Tiny and readable codebase.

---

## Installation

```
$ go get github.com/philippta/trip@latest
```

## Basic usage

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

            // Retry
            trip.Retry(attempts, delay),
            trip.Retry(attempts, delay, http.StatusInternalServerError, http.StatusBadGateway),
            trip.Retry(attempts, delay, trip.RetryableStatusCodes...),
        ),
    }

    client.Get("https://api.example.com/endpoint")
}
```

## API Overview

### Initialization

`trip.Default` creates a new transport, that wraps `http.DefaultTransport`. The transport can be then assigned to the transport of a new http.Client or as an override for the `http.DefaultTransport` itself. `trip.Default(...)` is the same as `trip.New(nil, ...)`.
```go
trip.Default(
    trip.BearerToken(apiToken),
    trip.Retry(attempts, delay, statusCodes...),
)
```

`trip.New` creates a new transport similar to `trip.Default()`, but let's you specify the underlying transport to wrap, if you already have one. Typically you would use `trip.Default(...)`.

```go
trip.New(http.DefaultTransport,
    trip.BearerToken(apiToken),
    trip.Retry(attempts, delay, statusCodes...),
)
```

### Bearer Token

`trip.BearerToken` sets the `Authorization` header to `Bearer <token>` on every request. Useful if you have an API token for an external service.

```go
trip.BearerToken(token)
```

### Basic Auth

`trip.BasicAuth` sets the `Authorization` header to `Basic <encoded-username-and-password>` on every request. Username and password are encoded according to RFC 7617.

```go
trip.BasicAuth(username, password)
```

### Retry

`trip.Retry` retries a request if either a network related issue is encountered or if the status code of the HTTP response matches any of the provided codes. `trip.RetryableStatusCodes` is a list of common HTTP status codes that can be retried.

```go
trip.Retry(attempts, delay)
trip.Retry(attempts, delay, statusCodes...)
trip.Retry(attempts, delay, trip.RetryableStatusCodes...)
```

### List of retryable HTTP status codes

`trip.RetryableStatusCodes` is a list of common HTTP status codes that are considered temporary and can be retried.

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

## Extending the default HTTP client

You can extend the transport of the default HTTP client (`http.DefaultClient`) for a quick and easy way to add authorization headers or retry behaviour to any of the default HTTP functions.

```go
http.DefaultClient.Transport = trip.Default(
    trip.BearerToken("api-token"),
    trip.Retry(3, 50 * time.Millisecond),
)

http.Get("https://api.example.com/endpoint")
http.Post("https://api.example.com/endpoint", "application/json", body)
http.PostForm("https://api.example.com/endpoint", fromData)
```


## Extending the default HTTP transport

You can extend the default transport (`http.DefaultTransport`) to add authorization headers or retry behaviour to any `http.Client` that is created throughout the livetime of your application.

⚠️ Be careful with this as this can alter the behavior of 3rd party libraries.

```go
http.DefaultTransport = trip.Default(
    trip.BearerToken("api-token"),
    trip.Retry(3, 50 * time.Millisecond),
)

client := &http.Client{} // uses http.DefaultTransport
client.Get("https://example.com/endpoint")
```
