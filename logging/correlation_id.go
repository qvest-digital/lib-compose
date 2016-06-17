package logging

import (
	"math/rand"
	"net/http"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var CorrelationIdHeader = "X-Correlation-Id"
var UserCorrelationCookie = ""

// EnsureCorrelationId returns the correlation from of the request.
// If the request does not have a correlation id, one will be generated and set to the request.
func EnsureCorrelationId(r *http.Request) string {
	id := r.Header.Get(CorrelationIdHeader)
	if id == "" {
		id = randStringBytes(10)
		r.Header.Set(CorrelationIdHeader, id)
	}
	return id
}

// GetCorrelationId returns the correlation from of the request.
func GetCorrelationId(r *http.Request) string {
	return r.Header.Get(CorrelationIdHeader)
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GetCorrelationId returns the correlation from of the request.
func GetUserCorrelationId(r *http.Request) string {
	if UserCorrelationCookie != "" {
		c, err := r.Cookie(UserCorrelationCookie)
		if err == nil {
			return c.Value
		}
	}
	return ""
}
