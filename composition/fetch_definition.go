package composition

import (
	"github.com/tarent/lib-compose/cache"
	"io"
	"net/http"
	"strings"
	"time"
)

// ForwardRequestHeaders are those headers,
// which are incuded from the original client request to the backend request.
// TODO: Add Host header to an XFF header
var ForwardRequestHeaders = []string{
	"Authorization",
	"Cache-Control",
	"Cookie",
	"Content-Length",
	"Content-Type",
	"If-Match",
	"If-Modified-Since",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
	"Pragma",
	"Referer",
	"Transfer-Encoding"}

// ForwardResponseHeaders are those headers,
// which are incuded from the servers backend response to the client.
var ForwardResponseHeaders = []string{
	"Age",
	"Allow",
	"Cache-Control",
	"Content-Disposition",
	"Content-Security-Policy",
	"Content-Type",
	"Date",
	"ETag",
	"Expires",
	"Last-Modified",
	"Link",
	"Location",
	"Pragma",
	"Set-Cookie",
	"WWW-Authenticate"}

const (
	DefaultTimeout time.Duration = 10 * time.Second
	FileURLPrefix                = "file://"
)

// FetchDefinition is a descriptor for fetching Content from an endpoint.
type FetchDefinition struct {
	URL           string
	Timeout       time.Duration
	Required      bool
	Header        http.Header
	Method        string
	Body          io.Reader
	RespProc      ResponseProcessor
	ErrHandler    ErrorHandler
	CacheStrategy CacheStrategy
	//ServeResponseHeaders bool
	//IsPrimary            bool
	//FallbackURL string
}

// Creates a fetch definition
func NewFetchDefinition(url string) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessor(url, nil)
}

func NewFetchDefinitionWithErrorHandler(url string, errHandler ErrorHandler) *FetchDefinition {
	if errHandler == nil {
		errHandler = NewDefaultErrorHandler()
	}
	return &FetchDefinition{
		URL:           url,
		Timeout:       DefaultTimeout,
		Required:      true,
		Method:        "GET",
		ErrHandler:    errHandler,
		CacheStrategy: cache.DefaultCacheStrategy,
	}
}

// If a ResponseProcessor-Implementation is given it can be used to change the response before composition
func NewFetchDefinitionWithResponseProcessor(url string, rp ResponseProcessor) *FetchDefinition {
	return &FetchDefinition{
		URL:           url,
		Timeout:       DefaultTimeout,
		Required:      true,
		Method:        "GET",
		RespProc:      rp,
		ErrHandler:    NewDefaultErrorHandler(),
		CacheStrategy: cache.DefaultCacheStrategy,
	}
}

// NewFetchDefinitionFromRequest creates a fetch definition
// from the request path, but replaces the sheme, host and port with the provided base url
func NewFetchDefinitionFromRequest(baseUrl string, r *http.Request) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorFromRequest(baseUrl, r, nil)
}

// NewFetchDefinitionFromRequest creates a fetch definition
// from the request path, but replaces the sheme, host and port with the provided base url
// If a ResponseProcessor-Implementation is given it can be used to change the response before composition
// Only those headers, defined in ForwardRequestHeaders are copied to the FetchDefinition.
func NewFetchDefinitionWithResponseProcessorFromRequest(baseUrl string, r *http.Request, rp ResponseProcessor) *FetchDefinition {
	if strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl[:len(baseUrl)-1]
	}

	fullPath := r.URL.Path
	if fullPath == "" {
		fullPath = "/"
	}
	if r.URL.RawQuery != "" {
		fullPath += "?" + r.URL.RawQuery
	}

	return &FetchDefinition{
		URL:        baseUrl + fullPath,
		Timeout:    DefaultTimeout,
		Header:     copyHeaders(r.Header, nil, ForwardRequestHeaders),
		Method:     r.Method,
		Body:       r.Body,
		Required:   true,
		RespProc:   rp,
		ErrHandler: NewDefaultErrorHandler(),
	}
}

// Hash returns a unique hash for the fetch request.
// If two hashes of fetch resources are equal, they refer the same resource
// and can e.g. be taken as replacement for each other. E.g. in case of caching.
func (def *FetchDefinition) Hash() string {
	if def.CacheStrategy != nil {
		return def.CacheStrategy.Hash(def.Method, def.URL, def.Header)
	}
	return def.URL
}

func (def *FetchDefinition) IsCachable(responseStatus int, responseHeaders http.Header) bool {
	if def.CacheStrategy != nil {
		return def.CacheStrategy.IsCachable(def.Method, def.URL, responseStatus, def.Header, responseHeaders)
	}
	return false
}

// copyHeaders copies only the header contained in the the whitelist
// from src to test. If dest is nil, it will be created.
// The dest will also be returned.
func copyHeaders(src, dest http.Header, whitelist []string) http.Header {
	if dest == nil {
		dest = http.Header{}
	}
	for _, k := range whitelist {
		headerValues := src[k]
		for _, v := range headerValues {
			dest.Add(k, v)
		}
	}
	return dest
}

// the default handler throws an status 502
type DefaultErrorHandler struct {
}

func NewDefaultErrorHandler() *DefaultErrorHandler {
	deh := new(DefaultErrorHandler)
	return deh
}

func (der *DefaultErrorHandler) Handle(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad Gateway: "+err.Error(), 502)
}
