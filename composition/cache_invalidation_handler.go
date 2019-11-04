package composition

import (
	"github.com/tarent/go-log-middleware/v2/logging"
	"net/http"
	"strings"
)

type CacheInvalidationHandler struct {
	cache Cache
	next  http.Handler
}

func (cih *CacheInvalidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" &&
		strings.Contains(r.URL.EscapedPath(), "internal/cache") &&
		cih.cache != nil {
		logging.Application(r.Header).Info("cache was invalidated")
		cih.cache.Invalidate()
	}
	if cih.next != nil {
		cih.next.ServeHTTP(w, r)
	}
}

func NewCacheInvalidationHandler(cache Cache, next http.Handler) *CacheInvalidationHandler {
	return &CacheInvalidationHandler{cache: cache, next: next}
}
