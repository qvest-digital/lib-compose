package util

import (
	"net/http"
)

const X_FORWARDED_HOST_HEADER_KEY = "X-Forwarded-Host"

type ForwardedHostHandler struct {
	Next http.Handler
}

func NewForwardedHostHandler(next http.Handler) *ForwardedHostHandler {
	return &ForwardedHostHandler{Next: next}
}

func (p *ForwardedHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Set(X_FORWARDED_HOST_HEADER_KEY, r.Host)

	if p.Next != nil {
		p.Next.ServeHTTP(w, r)
	}
}
