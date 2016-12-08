package cache

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/tarent/lib-compose/logging"
	"net/http"
	"strings"
)

const (
	// The request method was POST and an Expiration header was not supplied.
	ReasonRequestMethodPOST = cacheobject.ReasonRequestMethodPOST

	// The request method was PUT and PUTs are not cachable.
	ReasonRequestMethodPUT = cacheobject.ReasonRequestMethodPUT

	// The request method was DELETE and DELETEs are not cachable.
	ReasonRequestMethodDELETE = cacheobject.ReasonRequestMethodDELETE

	// The request method was CONNECT and CONNECTs are not cachable.
	ReasonRequestMethodCONNECT = cacheobject.ReasonRequestMethodCONNECT

	// The request method was OPTIONS and OPTIONS are not cachable.
	ReasonRequestMethodOPTIONS = cacheobject.ReasonRequestMethodOPTIONS

	// The request method was TRACE and TRACEs are not cachable.
	ReasonRequestMethodTRACE = cacheobject.ReasonRequestMethodTRACE

	// The request method was not recognized by cachecontrol, and should not be cached.
	ReasonRequestMethodUnkown = cacheobject.ReasonRequestMethodUnkown

	// The request included an Cache-Control: no-store header
	ReasonRequestNoStore = cacheobject.ReasonRequestNoStore

	// The request included an Authorization header without an explicit Public or Expiration time: http://tools.ietf.org/html/rfc7234#section-3.2
	ReasonRequestAuthorizationHeader = cacheobject.ReasonRequestAuthorizationHeader

	// The response included an Cache-Control: no-store header
	ReasonResponseNoStore = cacheobject.ReasonResponseNoStore

	// The response included an Cache-Control: private header and this is not a Private cache
	ReasonResponsePrivate = cacheobject.ReasonResponsePrivate

	// The response failed to meet at least one of the conditions specified in RFC 7234 section 3: http://tools.ietf.org/html/rfc7234#section-3
	ReasonResponseUncachableByDefault = cacheobject.ReasonResponseUncachableByDefault
)

var DefaultIncludeHeaders = []string{"Authorization", "Accept-Encoding", "Host"}

var DefaultCacheStrategy = NewCacheStrategyWithDefault()

type CacheStrategy struct {
	includeHeaders []string
	includeCookies []string
	ignoreReasons  []cacheobject.Reason
}

func NewCacheStrategyWithDefault() *CacheStrategy {
	return &CacheStrategy{
		includeHeaders: DefaultIncludeHeaders,
		includeCookies: nil,
		ignoreReasons:  nil,
	}
}

func NewCacheStrategy(includeHeaders []string, includeCookies []string, ignoreReasons []cacheobject.Reason) *CacheStrategy {
	return &CacheStrategy{
		includeHeaders: includeHeaders,
		includeCookies: includeCookies,
		ignoreReasons:  ignoreReasons,
	}
}

// Hash computes a hash value based on the url, the method and selected header and cookie attributes.
func (tcs *CacheStrategy) Hash(method string, url string, requestHeader http.Header) string {
	return tcs.HashWithParameters(method, url, requestHeader, tcs.includeHeaders, tcs.includeCookies)
}

// Hash computes a hash value based on the url, the method and selected header and cookie attributes.
func (tcs *CacheStrategy) HashWithParameters(method string, url string, requestHeader http.Header, includeHeaders []string, includeCookies []string) string {
	hasher := md5.New()

	hasher.Write([]byte(method))
	hasher.Write([]byte(url))

	for _, h := range includeHeaders {
		if requestHeader.Get(h) != "" {
			hasher.Write([]byte(h))
			hasher.Write([]byte(requestHeader.Get(h)))
		}
	}

	for _, c := range includeCookies {
		if value, found := readCookieValue(requestHeader, c); found {
			hasher.Write([]byte(c))
			hasher.Write([]byte(value))
		}
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
//	logging.Logger.Error("HASH = " + hash + ", url = " + url + ", host = " + requestHeader.Get("Host"))
	return hash
}

func (tcs *CacheStrategy) IsCacheable(method string, url string, statusCode int, requestHeader http.Header, responseHeader http.Header) bool {
	// TODO: it is expensive to create a request object only for passing to the cachecontrol library
	req := &http.Request{Method: method, Header: requestHeader}
	reasons, _, err := cacheobject.UsingRequestResponse(req, statusCode, responseHeader, true)
	if err != nil {
		logging.Logger.WithError(err).Warnf("error checking cachability for %v %v: %v", method, url, err)
		return false
	}
	for _, foundReason := range reasons {
		if !tcs.isReasonIgnorable(foundReason) {
			logging.Logger.WithField("notCachableReason", foundReason).
				WithField("type", "cacheinfo").
				Debugf("ressource not cachable %v %v: %v", method, url, foundReason)
			return false
		}
	}
	return true
}

func (tcs *CacheStrategy) isReasonIgnorable(reason cacheobject.Reason) bool {
	for _, ignoreReason := range tcs.ignoreReasons {
		if reason == ignoreReason {
			return true
		}
	}
	return false
}

// taken and adapted from net/http
func readCookieValue(h http.Header, filterName string) (string, bool) {
	lines, ok := h["Cookie"]
	if !ok {
		return "", false
	}

	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		for i := 0; i < len(parts); i++ {
			parts[i] = strings.TrimSpace(parts[i])
			if len(parts[i]) == 0 {
				continue
			}
			name, val := parts[i], ""
			if j := strings.Index(name, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}
			if filterName == name {
				if len(val) > 1 && val[0] == '"' && val[len(val)-1] == '"' {
					val = val[1 : len(val)-1]
				}
				return val, true
			}
		}
	}
	return "", false
}
