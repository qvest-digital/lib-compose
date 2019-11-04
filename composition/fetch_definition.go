package composition

import (
	"github.com/tarent/lib-compose/v2/cache"
	"github.com/tarent/lib-servicediscovery/servicediscovery"
	"io"
	"net/http"
	"strings"
	"time"
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
	"Host",
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
	// The name of the fetch definition
	Name                   string
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

// Creates a fetch definition (warning: this one will not forward any request headers).
func NewFetchDefinition(url string) *FetchDefinition {
	return &FetchDefinition{
		Name:            urlToName(url), // the name defauls to the url
		URL:             url,
		Timeout:         DefaultTimeout,
		FollowRedirects: false,
		Required:        true,
		Method:          "GET",
		ErrHandler:      NewDefaultErrorHandler(),
		CacheStrategy:   cache.DefaultCacheStrategy,
		Priority:        DefaultPriority,
	}
}

// Priority is used to determine which property from which head has to be taken by collision of multiple fetches
func (fd *FetchDefinition) WithPriority(priority int) *FetchDefinition {
	fd.Priority = priority
	return fd
}

// Use a given request to extract a path, method and body for the fetch request
func (fd *FetchDefinition) FromRequest(r *http.Request) *FetchDefinition {
	if strings.HasSuffix(fd.URL, "/") {
		fd.URL = fd.URL[:len(fd.URL)-1]
	}

	fullPath := r.URL.Path
	if fullPath == "" {
		fullPath = "/"
	}
	if r.URL.RawQuery != "" {
		fullPath += "?" + r.URL.RawQuery
	}

	fd.URL = fd.URL + fullPath
	fd.Body = r.Body
	fd.Method = r.Method
	fd.Header = copyHeaders(r.Header, fd.Header, ForwardRequestHeaders)

	return fd
}

// Copy headers to the fetchdefinition (but only the ones which are part of the whitelist)
func (fd *FetchDefinition) WithHeaders(header http.Header) *FetchDefinition {
	fd.Header = copyHeaders(header, fd.Header, ForwardRequestHeaders)
	return fd
}

// If a ResponseProcessor-Implementation is given it can be used to change the response before composition
func (fd *FetchDefinition) WithResponseProcessor(rp ResponseProcessor) *FetchDefinition {
	fd.RespProc = rp
	return fd
}

// Set a name to be used in the merge context later on
func (fd *FetchDefinition) WithName(name string) *FetchDefinition {
	fd.Name = name
	return fd
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

// Returns a name from a url, which has template placeholders eliminated
func urlToName(url string) string {
	url = strings.Replace(url, `§[`, `\§\[`, -1)
	url = strings.Replace(url, `]§`, `\]\§`, -1)
	return url
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
