package composition

import "net/http"

type CacheInvalidationHandler struct {
	Cache Cache
	Next  http.Handler
}

func (cih *CacheInvalidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cih.Cache.Invalidate()
	if cih.Next != nil {
		cih.Next.ServeHTTP(w, r)
	}
}
