package composition

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_CompositionHandler_PositiveCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	content := &MemoryContent{
		body: map[string]Fragment{
			"": StringFragment("Hello World\n"),
		},
	}

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			&FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: content,
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory), nil)

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	aggregator.ServeHTTP(resp, r)

	expected := `<html>
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
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory), nil)
	aggregator.contentMergerFactory = func() ContentMerger {
		merger := NewMockContentMerger(ctrl)
		merger.EXPECT().AddContent(gomock.Any())
		merger.EXPECT().AddMetaValue(gomock.Any(), gomock.Any())
		merger.EXPECT().WriteHtml(gomock.Any()).Return(errors.New("an error"))
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
				Content: nil,
				Err:     errors.New(errorString),
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory), nil)

	resp := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	aggregator.ServeHTTP(resp, r)

	a.Equal("Bad Gateway: "+errorString+"\n", string(resp.Body.Bytes()))
	a.Equal(502, resp.Code)
}

func Test_metadataForReqest(t *testing.T) {
	a := assert.New(t)
	r, _ := http.NewRequest("GET", "https://example.com/nothing?foo=bar", nil)

	m := metadataForReqest(r)
	a.Equal("http://example.com", m["base_url"])
	a.Equal("bar", m["params"].(url.Values).Get("foo"))
}

type MockFetchResultSupplier []*FetchResult

func (m MockFetchResultSupplier) WaitForResults() []*FetchResult {
	return []*FetchResult(m)
}
