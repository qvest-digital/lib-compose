package composition

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/lib-compose/cache"
	"golang.org/x/net/html"
)

func Test_CompositionHandler_PositiveCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		frag := NewStringFragment("Hello World\n")
		frag.AddLinkTags([][]html.Attribute{stylesheetAttrs("/path/to/style1.css"),
			stylesheetAttrs("/path/to/style2.css"),
			stylesheetAttrs("/path/to/style1.css"),
			stylesheetAttrs("/path/to/style2.css")})

		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": frag,
					},
				},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	ch.ServeHTTP(resp, r)

	expected := `<!DOCTYPE html>
<html>
  <head>
    
    <link rel="stylesheet" type="text/css" href="/path/to/style1.css">
    <link rel="stylesheet" type="text/css" href="/path/to/style2.css">
    <link rel="stylesheet" type="text/css" href="/path/to/style1.css">
    <link rel="stylesheet" type="text/css" href="/path/to/style2.css">
  </head>
  <body>
    Hello World

  </body>
</html>
`
	a.Equal(expected, string(resp.Body.Bytes()))
	a.Equal(200, resp.Code)
}

func Test_CompositionHandler_PositiveCaseWithSimpleDeduplicationStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		frag := NewStringFragment("Hello World\n")
		frag.AddLinkTags([][]html.Attribute{stylesheetAttrs("/path/to/style1.css"),
			stylesheetAttrs("/path/to/style2.css"),
			stylesheetAttrs("/path/to/style1.css"),
			stylesheetAttrs("/path/to/style2.css")})

		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": frag,
					},
				},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))
	factory := func() DeduplicationStrategy {
		return new(SimpleDeduplicationStrategy)
	}
	ch.WithDeduplicationStrategyFactory(factory)

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	ch.ServeHTTP(resp, r)

	expected := `<!DOCTYPE html>
<html>
  <head>
    
    <link rel="stylesheet" type="text/css" href="/path/to/style1.css">
    <link rel="stylesheet" type="text/css" href="/path/to/style2.css">
  </head>
  <body>
    Hello World

  </body>
</html>
`
	a.Equal(expected, string(resp.Body.Bytes()))
	a.Equal(200, resp.Code)
}

func Test_CompositionHandler_PositiveCaseWithCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": NewStringFragment("Hello World\n"),
					},
				},
				Hash: "hashString",
			},
		}
	}
	ch := NewCompositionHandlerWithCache(ContentFetcherFactory(contentFetcherFactory), cache.NewCache("my-cache", 100, 100, time.Millisecond))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	ch.cache.Set("hashString", "", 1, nil)
	ch.ServeHTTP(resp, r)

	_, foundInCache := ch.cache.Get("hashString")
	a.True(foundInCache)

}

func Test_CompositionHandler_CorrectHeaderAndStatusCodeReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": NewStringFragment(""),
					},
					httpHeader: http.Header{
						"Transfer-Encoding": {"gzip"}, // removed
						"Set-Cookie": {
							"cookie-content 1",
							"cookie-content 2",
						},
					},
					httpStatusCode: 200,
				},
			},
			&FetchResult{
				Def: NewFetchDefinition("..."),
				Content: &MemoryContent{
					httpHeader: http.Header{
						"Set-Cookie": {
							"cookie-content 3",
						},
					},
					httpStatusCode: 201,
				},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	ch.ServeHTTP(resp, r)

	a.Equal(200, resp.Code)
	a.Equal(3, len(resp.Header())) // Set-Cookie + Content-Type + Content-Length
	a.Equal("", resp.Header().Get("Transfer-Encoding"))
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 1")
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 2")
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 3")
}

func Test_CompositionHandler_CorrectHeaderAndStatusCodeReturned_onRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": NewStringFragment(""),
					},
					httpHeader: http.Header{
						"Transfer-Encoding": {"gzip"}, // removed
						"Location":          {"/look/somewhere"},
						"Set-Cookie": {
							"cookie-content 1",
							"cookie-content 2",
						},
					},
					httpStatusCode: 302,
				},
			},
			&FetchResult{
				Def:     NewFetchDefinition("..."),
				Content: &MemoryContent{},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	ch.ServeHTTP(resp, r)

	a.Equal(302, resp.Code)
	a.Equal(2, len(resp.Header()))
	a.Equal("/look/somewhere", resp.Header().Get("Location"))
	a.Equal("", resp.Header().Get("Transfer-Encoding"))
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 1")
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 2")
}

func Test_CompositionHandler_HeadRequest_CorrectHeaderAndStatusCodeReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					httpHeader: http.Header{
						"Transfer-Encoding": {"gzip"}, // removed
						"Location":          {"/look/somewhere"},
						"Set-Cookie": {
							"cookie-content 1",
							"cookie-content 2",
						},
					},
					httpStatusCode: 200,
				},
			},
			&FetchResult{
				Def:     NewFetchDefinition("..."),
				Content: &MemoryContent{},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("HEAD", "http://example.com", nil)
	ch.ServeHTTP(resp, r)

	a.Equal(200, resp.Code)
	a.Equal(2, len(resp.Header()))
	a.Equal("/look/somewhere", resp.Header().Get("Location"))
	a.Equal("", resp.Header().Get("Transfer-Encoding"))
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 1")
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 2")
}

func Test_CompositionHandler_CorrectStatusCodeReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": NewStringFragment(""),
					},
					httpHeader: http.Header{
						"Transfer-Encoding": {"gzip"}, // removed
						"Location":          {"/look/somewhere"},
						"Set-Cookie": {
							"cookie-content 1",
							"cookie-content 2",
						},
					},
					httpStatusCode: 200,
				},
			},
			&FetchResult{
				Def:     NewFetchDefinition("..."),
				Content: &MemoryContent{},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	ch.ServeHTTP(resp, r)

	a.Equal(200, resp.Code)
	a.Equal(4, len(resp.Header()))
	a.Equal("/look/somewhere", resp.Header().Get("Location"))
	a.Equal("", resp.Header().Get("Transfer-Encoding"))
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 1")
	a.Contains(resp.Header()["Set-Cookie"], "cookie-content 2")
}

func Test_CompositionHandler_ReturnStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentWithFragment := &MemoryContent{
		body: map[string]Fragment{
			"": NewStringFragment("Hello World\n"),
		},
	}

	contentWithReader := &MemoryContent{
		reader:         ioutil.NopCloser(strings.NewReader("bar")),
		httpHeader:     http.Header{"Content-Type": {"text/plain"}},
		httpStatusCode: 201,
	}

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def:     NewFetchDefinition("/fragment"),
				Content: contentWithFragment,
			},
			&FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: contentWithReader,
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	aggregator.ServeHTTP(resp, r)

	a.Equal("bar", string(resp.Body.Bytes()))
	a.Equal("text/plain", resp.Header().Get("Content-Type"))
	a.Equal(201, resp.Code)
}

func Test_CompositionHandler_ErrorInMerging(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: &MemoryContent{},
				Err:     nil,
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))
	aggregator.contentMergerFactory = func(jsonData map[string]interface{}) ContentMerger {
		merger := NewMockContentMerger(ctrl)
		merger.EXPECT().AddContent(gomock.Any(), 0)
		merger.EXPECT().GetHtml().Return(nil, errors.New("an error"))
		return merger
	}

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)

	aggregator.ServeHTTP(resp, r)

	a.Equal("Internal Server Error: an error\n", string(resp.Body.Bytes()))
	a.Equal(500, resp.Code)
}

func Test_CompositionHandler_ErrorInMergingWithCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: &MemoryContent{},
				Err:     nil,
				Hash:    "hashString",
			},
		}
	}

	aggregator := NewCompositionHandlerWithCache(ContentFetcherFactory(contentFetcherFactory), cache.NewCache("my-cache", 100, 100, time.Millisecond))
	aggregator.cache.Set("hashString", "", 1, nil)
	aggregator.contentMergerFactory = func(jsonData map[string]interface{}) ContentMerger {
		merger := NewMockContentMerger(ctrl)
		merger.EXPECT().AddContent(gomock.Any(), 0)
		merger.EXPECT().GetHtml().Return(nil, errors.New("an error"))
		return merger
	}

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)

	aggregator.ServeHTTP(resp, r)

	_, foundInCache := aggregator.cache.Get("hashString")

	a.False(foundInCache)
	a.Equal("Internal Server Error: an error\n", string(resp.Body.Bytes()))
	a.Equal(500, resp.Code)
}

func Test_CompositionHandler_ErrorInFetching(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	errorString := "some error while fetching"
	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: &MemoryContent{httpStatusCode: 502},
				Err:     errors.New(errorString),
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	aggregator.ServeHTTP(resp, r)

	a.Equal("Error: "+errorString+"\n", string(resp.Body.Bytes()))
	a.Equal(502, resp.Code)
}

func Test_CompositionHandler_ErrorEmptyFetchersList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	aggregator.ServeHTTP(resp, r)

	a.Equal("Internal server error", string(resp.Body.Bytes()))
	a.Equal(500, resp.Code)
}

func Test_metadataForRequest(t *testing.T) {
	a := assert.New(t)
	r, _ := http.NewRequest("GET", "https://example.com/nothing?foo=bar", nil)

	m := MetadataForRequest(r)
	a.Equal("http://example.com", m["base_url"])
	a.Equal("example.com", m["host"])
	a.Equal("bar", m["params"].(url.Values).Get("foo"))
}

func Test_getBaseUrlFromRequest(t *testing.T) {
	a := assert.New(t)

	tests := []struct {
		rURL         string
		headers      http.Header
		tls          bool
		expectedURL  string
		expectedHost string
	}{
		{
			rURL:         "http://example.com/nothing?foo=bar",
			expectedURL:  "http://example.com",
			expectedHost: "example.com",
		},
		{
			rURL:         "http://example.com:8080/sdcsd",
			expectedURL:  "http://example.com:8080",
			expectedHost: "example.com:8080",
		},
		{
			rURL:         "http://example.com/nothing?foo=bar",
			tls:          true,
			expectedURL:  "https://example.com",
			expectedHost: "example.com",
		},
		{
			rURL: "http://example.com/nothing?foo=bar",
			headers: http.Header{"X-Forwarded-For": {"other.com"},
				"X-Forwarded-Proto": {"https"}},
			expectedURL:  "https://other.com",
			expectedHost: "other.com",
		},
		{
			rURL: "http://example.com/nothing?foo=bar",
			headers: http.Header{"X-Forwarded-For": {"other.com, xyz"},
				"X-Forwarded-Proto": {"https, xyz"}},
			expectedURL:  "https://other.com",
			expectedHost: "other.com",
		},
	}
	for _, test := range tests {
		r, _ := http.NewRequest("GET", test.rURL, nil)
		if test.tls {
			r.TLS = &tls.ConnectionState{}
		}
		if test.headers != nil {
			r.Header = test.headers
		}
		url := getBaseUrlFromRequest(r)
		a.Equal(test.expectedURL, url)
		host := getHostFromRequest(r)
		a.Equal(test.expectedHost, host)
	}
}

// Jira 3946: go deletes the "Host" header from the request (for whatever reasons):
// https://golang.org/src/net/http/request.go:
//   123		// For incoming requests, the Host header is promoted to the
//   124		// Request.Host field and removed from the Header map.
// But our cache strategies might want this header in order to generate different
// cache-IDs (hashes) for different host names (e.g. preview-mode-whatever.de vs. production-whatever.de).
func Test_CompositionHandler_RestoreHostHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": NewStringFragment("Hello World\n"),
					},
				},
			},
		}
	}
	ch := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	r.Host = "MyHost"
	ch.ServeHTTP(resp, r)

	a.Equal(200, resp.Code)
	a.Equal("MyHost", r.Header.Get("Host"))
}

type MockFetchResultSupplier []*FetchResult

func (m MockFetchResultSupplier) WaitForResults() []*FetchResult {
	return []*FetchResult(m)
}

func (m MockFetchResultSupplier) MetaJSON() map[string]interface{} {
	return nil
}

func (m MockFetchResultSupplier) Empty() bool {
	return len([]*FetchResult(m)) == 0
}
