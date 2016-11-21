package composition

import (
	"github.com/tarent/lib-compose/cache"
	"github.com/tarent/lib-servicediscovery/servicediscovery"
	"io"
	"net/http"
	"strings"
	"time"
	"github.com/tarent/lib-compose/logging"
)

const MAX_PRIORITY int = 4294967295

// ForwardRequestHeaders are those headers,
// which are included from the original client request to the backend request.
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
	"Transfer-Encoding",
	"X-Forwarded-Host",
	"X-Correlation-Id",
	"X-Feature-Toggle",
}

// ForwardResponseHeaders are those headers,
// which are included from the servers backend response to the client.
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
	DefaultTimeout  time.Duration = 10 * time.Second
	FileURLPrefix                 = "file://"
	DefaultPriority               = 0
)

// FetchDefinition is a descriptor for fetching Content from an endpoint.
type FetchDefinition struct {
	URL                    string
	Timeout                time.Duration
	FollowRedirects        bool
	Required               bool
	Header                 http.Header
	Method                 string
	Body                   io.Reader
	RespProc               ResponseProcessor
	ErrHandler             ErrorHandler
	CacheStrategy          CacheStrategy
	ServiceDiscoveryActive bool
	ServiceDiscovery       servicediscovery.ServiceDiscovery
	Priority               int
}

// Creates a fetch definition
func NewFetchDefinition(url string) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorAndPriority(url, nil, DefaultPriority)
}

func NewFetchDefinitionWithPriority(url string, priority int) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorAndPriority(url, nil, priority)
}

func NewFetchDefinitionWithErrorHandler(url string, errHandler ErrorHandler) *FetchDefinition {
	return NewFetchDefinitionWithErrorHandlerAndPriority(url, errHandler, DefaultPriority)
}

func NewFetchDefinitionWithErrorHandlerAndPriority(url string, errHandler ErrorHandler, priority int) *FetchDefinition {
	if errHandler == nil {
		errHandler = NewDefaultErrorHandler()
	}
	return &FetchDefinition{
		URL:             url,
		Timeout:         DefaultTimeout,
		FollowRedirects: false,
		Required:        true,
		Method:          "GET",
		ErrHandler:      errHandler,
		CacheStrategy:   cache.DefaultCacheStrategy,
		Priority:        priority,
	}
}

// If a ResponseProcessor-Implementation is given it can be used to change the response before composition
func NewFetchDefinitionWithResponseProcessor(url string, rp ResponseProcessor) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorAndPriority(url, rp, DefaultPriority)
}

// If a ResponseProcessor-Implementation is given it can be used to change the response before composition
// Priority is used to determine which property from which head has to be taken by collision of multiple fetches
func NewFetchDefinitionWithResponseProcessorAndPriority(url string, rp ResponseProcessor, priority int) *FetchDefinition {
	return &FetchDefinition{
		URL:             url,
		Timeout:         DefaultTimeout,
		FollowRedirects: false,
		Required:        true,
		Method:          "GET",
		RespProc:        rp,
		ErrHandler:      NewDefaultErrorHandler(),
		CacheStrategy:   cache.DefaultCacheStrategy,
		Priority:        priority,
	}
}

// NewFetchDefinitionFromRequest creates a fetch definition
// from the request path, but replaces the scheme, host and port with the provided base url
func NewFetchDefinitionFromRequest(baseUrl string, r *http.Request) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorAndPriorityFromRequest(baseUrl, r, nil, DefaultPriority)
}

// NewFetchDefinitionFromRequest creates a fetch definition
// from the request path, but replaces the scheme, host and port with the provided base url
func NewFetchDefinitionWithPriorityFromRequest(baseUrl string, r *http.Request, priority int) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorAndPriorityFromRequest(baseUrl, r, nil, priority)
}

// NewFetchDefinitionFromRequest creates a fetch definition
// from the request path, but replaces the scheme, host and port with the provided base url
// If a ResponseProcessor-Implementation is given it can be used to change the response before composition
// Only those headers, defined in ForwardRequestHeaders are copied to the FetchDefinition.
func NewFetchDefinitionWithResponseProcessorFromRequest(baseUrl string, r *http.Request, rp ResponseProcessor) *FetchDefinition {
	return NewFetchDefinitionWithResponseProcessorAndPriorityFromRequest(baseUrl, r, rp, DefaultPriority)
}

// NewFetchDefinitionWithResponseProcessorFromRequest with priority setting for head property collision handling
func NewFetchDefinitionWithResponseProcessorAndPriorityFromRequest(baseUrl string, r *http.Request, rp ResponseProcessor, priority int) *FetchDefinition {
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
		URL:             baseUrl + fullPath,
		Timeout:         DefaultTimeout,
		FollowRedirects: false,
		Header:          copyHeaders(r.Header, nil, ForwardRequestHeaders),
		Method:          r.Method,
		Body:            r.Body,
		Required:        true,
		RespProc:        rp,
		ErrHandler:      NewDefaultErrorHandler(),
		Priority:        priority,
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

func (def *FetchDefinition) IsCacheable(responseStatus int, responseHeaders http.Header) bool {
	if def.CacheStrategy != nil {
		return def.CacheStrategy.IsCacheable(def.Method, def.URL, responseStatus, def.Header, responseHeaders)
	}
	return false
}

func (def *FetchDefinition) IsReadableFromCache() bool {
	return def.IsCacheable(200, nil)
}

// copyHeaders copies only the header contained in the the whitelist
// from src to dest. If dest is nil, it will be created.
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

	//Set the correlation-id in the http header, so the services will retrieve and use it for further logging
        dest.Add("correlation_id", logging.GetCorrelationId(dest))

	return dest
}

// the default handler throws an status 502
type DefaultErrorHandler struct {
}

func NewDefaultErrorHandler() *DefaultErrorHandler {
	deh := new(DefaultErrorHandler)
	return deh
}

func (der *DefaultErrorHandler) Handle(err error, status int, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Error: "+err.Error(), status)
}
