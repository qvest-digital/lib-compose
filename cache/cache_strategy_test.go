package cache

import (
	"fmt"
	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/lib-compose/util"
	"net/http"
	"testing"
)

type hashCall struct {
	method        string
	url           string
	requestHeader http.Header
}

func Test_CacheStrategy_FetchDefinitionHash(t *testing.T) {
	a := assert.New(t)
	tests := []struct {
		strategy *CacheStrategy
		call1    hashCall
		call2    hashCall
		equal    bool
	}{
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", nil},
			hashCall{"GET", "/foo", nil},
			true,
		},
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", http.Header{"Accept": {"text/html"}}},
			hashCall{"GET", "/foo", http.Header{}},
			true,
		},
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", http.Header{"Cookie": {"foo=bar; bi=bo"}}},
			hashCall{"GET", "/foo", http.Header{}},
			true,
		},
		{
			NewCacheStrategy(nil, []string{"foo"}, nil),
			hashCall{"GET", "/foo", http.Header{"Cookie": {"cname=bar; bi=bo"}}},
			hashCall{"GET", "/foo", http.Header{}},
			true,
		},
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", nil},
			hashCall{"GET", "/bar", nil},
			false,
		},
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", nil},
			hashCall{"POST", "/foo", nil},
			false,
		},
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", http.Header{"Authorization": {"Basic: foo"}}},
			hashCall{"GET", "/foo", http.Header{"Authorization": {"Basic: bar"}}},
			false,
		},
		{
			DefaultCacheStrategy,
			hashCall{"GET", "/foo", http.Header{"Accept-Encoding": {"gzip"}}},
			hashCall{"GET", "/foo", http.Header{}},
			false,
		},
		{
			NewCacheStrategy(nil, []string{"cname"}, nil),
			hashCall{"GET", "/foo", http.Header{"Cookie": {"cname=bar"}}},
			hashCall{"GET", "/foo", http.Header{}},
			false,
		},
		{
			NewCacheStrategy([]string{"Accept"}, nil, nil),
			hashCall{"GET", "/foo", http.Header{"Accept": {"text/html"}}},
			hashCall{"GET", "/foo", http.Header{}},
			false,
		},
	}

	for _, t := range tests {
		hash1 := t.strategy.Hash(t.call1.method, t.call1.url, t.call1.requestHeader)
		hash2 := t.strategy.Hash(t.call2.method, t.call2.url, t.call2.requestHeader)
		if t.equal {
			a.Equal(hash1, hash2)
		} else {
			a.NotEqual(hash1, hash2)
		}
	}
}

func Test_CacheStrategy_IsCachable(t *testing.T) {
	a := assert.New(t)
	tests := []struct {
		strategy       *CacheStrategy
		method         string
		statusCode     int
		requestHeader  http.Header
		responseHeader http.Header
		isCacheable    bool
	}{
		{
			DefaultCacheStrategy,
			"GET",
			200,
			http.Header{
				"Accept":        {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
				"Cache-Control": {"max-age=0"},
				"Connection":    {"keep-alive"},
			},
			http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
				"Date":         {"Wed, 22 Jun 2016 12:03:25 GMT"},
			},
			true,
		},
		{
			DefaultCacheStrategy,
			"GET",
			200,
			nil,
			http.Header{
				"Authorization": {"Basic sdcsad"},
			},
			true,
		},
		{
			DefaultCacheStrategy,
			"POST",
			200,
			nil,
			nil,
			false,
		},
		{
			NewCacheStrategy(nil, nil, []cacheobject.Reason{cacheobject.ReasonRequestMethodPOST}),
			"POST",
			200,
			nil,
			nil,
			true,
		},
		{
			DefaultCacheStrategy,
			"GET",
			500,
			nil,
			nil,
			false,
		},
		{
			DefaultCacheStrategy,
			"GET",
			200,
			http.Header{"Cache-Control": {"max-age=-1"}}, // error case of lib cacheobject
			nil,
			false,
		},
		{
			DefaultCacheStrategy,
			"GET",
			200,
			nil,
			http.Header{
				"Cache-Control": {"no-store"},
			},
			false,
		},
		{
			DefaultCacheStrategy,
			"GET",
			200,
			nil,
			http.Header{
				"Cache-Control": {"no-store, no-cache"},
			},
			false,
		},
	}

	for _, t := range tests {
		cacheable := t.strategy.IsCacheable(t.method, "", t.statusCode, t.requestHeader, t.responseHeader)
		message := fmt.Sprintf("%v = isCacheable(%q, %q, %v, %v, %v)", cacheable, t.method, "", t.statusCode, t.requestHeader, t.responseHeader)
		if t.isCacheable {
			a.True(cacheable, message)
		} else {
			a.False(cacheable, message)
		}
	}
}

func Test_CacheStrategy_readCookieValue(t *testing.T) {
	a := assert.New(t)

	v, found := util.ReadCookieValue(http.Header{"Cookie": {"foo=bar"}}, "foo")
	a.True(found)
	a.Equal("bar", v)

	v, found = util.ReadCookieValue(http.Header{"Cookie": {`foo="bar"`}}, "foo")
	a.True(found)
	a.Equal("bar", v)

	v, found = util.ReadCookieValue(http.Header{"Cookie": {"foo"}}, "foo")
	a.True(found)
	a.Equal("", v)

	v, found = util.ReadCookieValue(http.Header{"Cookie": {";"}}, "foo")
	a.False(found)

	v, found = util.ReadCookieValue(http.Header{"Cookie": {""}}, "foo")
	a.False(found)
}
