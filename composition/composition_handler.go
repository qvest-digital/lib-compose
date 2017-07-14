package composition

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/tarent/lib-compose/logging"
)

// A ContentFetcherFactory returns a configured fetch job for a request
// which can return the fetch results.
type ContentFetcherFactory func(r *http.Request) FetchResultSupplier

type CompositionHandler struct {
	contentFetcherFactory ContentFetcherFactory
	contentMergerFactory  func(metaJSON map[string]interface{}) ContentMerger
	cache                 Cache
}

// NewCompositionHandler creates a new Handler with the supplied defaultData,
// which is used for each request.
func NewCompositionHandler(contentFetcherFactory ContentFetcherFactory) *CompositionHandler {
	return &CompositionHandler{
		contentFetcherFactory: contentFetcherFactory,
		contentMergerFactory: func(metaJSON map[string]interface{}) ContentMerger {
			return NewContentMerge(metaJSON)
		},
		cache: nil,
	}
}

// NewCompositionHandlerWithCache creates a new Handler with the supplied defaultData,
// which is used for each request.
// Use this constructor, if you created a caching content loader and provide
// the handle to it's cache as argument.
func NewCompositionHandlerWithCache(contentFetcherFactory ContentFetcherFactory, cache Cache) *CompositionHandler {
	return NewCompositionHandler(contentFetcherFactory).WithCache(cache)
}

func (agg *CompositionHandler) WithCache(cache Cache) *CompositionHandler {
	agg.cache = cache
	return agg
}

// Set the deduplication strategy to be used by the constructed content merger
// This method will first take effect in the upcomping call of ServeHTTP()
func (agg *CompositionHandler) WithDeduplicationStrategyFactory(strategyFactory func() DeduplicationStrategy) *CompositionHandler {
	wrapped := agg.contentMergerFactory
	agg.contentMergerFactory = func(metaJSON map[string]interface{}) ContentMerger {
		cm := wrapped(metaJSON)
		if cm != nil {
			cm.SetDeduplicationStrategy(strategyFactory())
		}
		return cm
	}
	return agg
}

func (agg *CompositionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If we know the host but don't have the Host header [any more] then we
	// set [or restore] the header, because why would You just remove it!?:
	if (r.Host != "") && (r.Header.Get("Host") == "") {
		r.Header.Set("Host", r.Host)
	}

	fetcher := agg.contentFetcherFactory(r)

	if agg.handleEmptyFetcher(fetcher, w, r) {
		return
	}

	// fetch all contents
	results := fetcher.WaitForResults()

	// Allow HEAD requests and disable composition of body fragments
	if agg.handleHeadRequests(results, w, r) {
		return
	}

	mergeContext := agg.contentMergerFactory(fetcher.MetaJSON())

	for _, res := range results {
		if res.Err == nil && res.Content != nil {

			// Handle responses with 30x status code or with response bodies
			if agg.handleNonMergeableResponses(res, w, r) {
				return
			}

			mergeContext.AddContent(res.Content, res.Def.Priority)

		} else if res.Def.Required {
			LogFetchResultLoadingError(res, w, r)
			return
		} else {
			logging.Application(r.Header).WithField("fetchResult", res).Warnf("optional content not loaded: %v", res.Def.URL)
		}
	}

	status := agg.extractStatusCode(results, w, r)

	agg.copyHeadersIfNeeded(results, w, r)

	// Overwrite Content-Type to ensure, that the encoding is correct
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html, err := agg.processHtml(mergeContext, w, r)
	// Return if an error occured within the html aggregation
	if err != nil {
		agg.purgeCacheEntries(results)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(html)))
	w.WriteHeader(status)
	w.Write(html)
}

// Purge the documents with the supplied hashes out of the cache.
func (agg *CompositionHandler) purgeCacheEntries(results []*FetchResult) {
	if agg.cache != nil {
		hashes := []string{}
		for _, r := range results {
			hashes = append(hashes, r.Hash)
		}
		agg.cache.PurgeEntries(hashes)
	}
}

func (agg *CompositionHandler) handleNonMergeableResponses(result *FetchResult, w http.ResponseWriter, r *http.Request) bool {
	if agg.handle30xResponses(result, w, r) {
		// Return if it's a forwarded status code
		return true
	}

	if agg.handleStreamResponses(result, w, r) {
		// Return if it's a response with body
		return true
	}

	return false
}

func (agg *CompositionHandler) extractStatusCode(results []*FetchResult, w http.ResponseWriter, r *http.Request) (statusCode int) {
	if len(results) > 0 {
		if results[0].Content.HttpStatusCode() != 0 {
			return results[0].Content.HttpStatusCode()
		}
	}
	return 200
}

func (agg *CompositionHandler) copyHeadersIfNeeded(results []*FetchResult, w http.ResponseWriter, r *http.Request) {
	// Take headers from first fetch definition
	if len(results) > 0 {
		copyHeaders(results[0].Content.HttpHeader(), w.Header(), ForwardResponseHeaders)
	}

	// But also allow results from other fetch definitions to set cookies, if Set-Cookie ist globaly allowed
	if len(results) > 1 && contains(ForwardResponseHeaders, "Set-Cookie") {
		for _, r := range results[1:] {
			copyHeaders(r.Content.HttpHeader(), w.Header(), []string{"Set-Cookie"})
		}
	}
}

func (agg *CompositionHandler) processHtml(mergeContext ContentMerger, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	html, err := mergeContext.GetHtml()
	if err != nil {
		logging.Application(r.Header).Error(err.Error())
		http.Error(w, "Internal Server Error: "+err.Error(), 500)
	}
	return html, err
}

func (agg *CompositionHandler) handleHeadRequests(results []*FetchResult, w http.ResponseWriter, r *http.Request) bool {
	if r.Method == "HEAD" && len(results) > 0 {
		copyHeaders(results[0].Content.HttpHeader(), w.Header(), ForwardResponseHeaders)
		w.WriteHeader(results[0].Content.HttpStatusCode())
		return true
	}
	return false
}

func (agg *CompositionHandler) handleEmptyFetcher(fetcher FetchResultSupplier, w http.ResponseWriter, r *http.Request) bool {
	if fetcher.Empty() {
		w.WriteHeader(500)
		w.Write([]byte("Internal server error"))
		logging.Application(r.Header).Error("No fetchers available for composition, throwing error 500")
		return true
	}
	return false
}

func (agg *CompositionHandler) handle30xResponses(result *FetchResult, w http.ResponseWriter, r *http.Request) bool {
	if result.Content.HttpStatusCode() >= 300 && result.Content.HttpStatusCode() <= 308 {
		copyHeaders(result.Content.HttpHeader(), w.Header(), ForwardResponseHeaders)
		w.WriteHeader(result.Content.HttpStatusCode())
		return true
	}
	return false
}

func (agg *CompositionHandler) handleStreamResponses(result *FetchResult, w http.ResponseWriter, r *http.Request) bool {
	if result.Content.Reader() != nil {
		copyHeaders(result.Content.HttpHeader(), w.Header(), ForwardResponseHeaders)
		w.WriteHeader(result.Content.HttpStatusCode())
		io.Copy(w, result.Content.Reader())
		result.Content.Reader().Close()
		return true
	}
	return false
}

func LogFetchResultLoadingError(res *FetchResult, w http.ResponseWriter, r *http.Request) {
	// 404 and 502 Error already become logged in logger.go
	if res.Content.HttpStatusCode() != 404 && res.Content.HttpStatusCode() != 502 {
		logging.Application(r.Header).WithField("fetchResult", res).Errorf("error loading content from: %v", res.Def.URL)
	}
	res.Def.ErrHandler.Handle(res.Err, res.Content.HttpStatusCode(), w, r)
}

func MetadataForRequest(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"host":     getHostFromRequest(r),
		"base_url": getBaseUrlFromRequest(r),
		"params":   r.URL.Query(),
	}
}

func getBaseUrlFromRequest(r *http.Request) string {
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if xfph := r.Header.Get("X-Forwarded-Proto"); xfph != "" {
		protoParts := strings.SplitN(xfph, ",", 2)
		proto = protoParts[0]
	}

	return proto + "://" + getHostFromRequest(r)
}

func getHostFromRequest(r *http.Request) string {
	host := r.Host
	if xffh := r.Header.Get("X-Forwarded-For"); xffh != "" {
		hostParts := strings.SplitN(xffh, ",", 2)
		host = hostParts[0]
	}
	return host
}

func hasPrioritySetting(results []*FetchResult) bool {
	for _, res := range results {
		if res.Def.Priority > 0 {
			return true
		}
	}
	return false
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}
