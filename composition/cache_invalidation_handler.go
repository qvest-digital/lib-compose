package composition

import "net/http"

type CacheInvalidationHandler struct {
	cache Cache
	next  http.Handler
}

func (cih *CacheInvalidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cih.cache.Invalidate()
	if cih.next != nil {
		cih.next.ServeHTTP(w, r)
	}
}

func NewCacheInvalidationHandler(cache Cache, next http.Handler) *CacheInvalidationHandler {
	return &CacheInvalidationHandler{cache: cache, next: next}
}
