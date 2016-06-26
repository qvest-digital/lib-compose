package composition

import (
	"crypto/tls"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_CompositionHandler_PositiveCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def: NewFetchDefinition("/foo"),
				Content: &MemoryContent{
					body: map[string]Fragment{
						"": StringFragment("Hello World\n"),
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
    
  </head>
  <body>
    Hello World

  </body>
</html>
`
	a.Equal(expected, string(resp.Body.Bytes()))
	a.Equal(200, resp.Code)
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
						"": StringFragment(""),
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

func Test_CompositionHandler_ReturnStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	contentWithFragment := &MemoryContent{
		body: map[string]Fragment{
			"": StringFragment("Hello World\n"),
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
		merger.EXPECT().AddContent(gomock.Any())
		merger.EXPECT().GetHtml().Return(nil, errors.New("an error"))
		return merger
	}

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	aggregator.ServeHTTP(resp, r)

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

type MockFetchResultSupplier []*FetchResult

func (m MockFetchResultSupplier) WaitForResults() []*FetchResult {
	return []*FetchResult(m)
}

func (m MockFetchResultSupplier) MetaJSON() map[string]interface{} {
	return nil
}
