package composition

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_CompositionHandler_PositiveCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	content := &MemoryContent{
		body: map[string]Fragment{
			"main": StringFragment("Hello World\n"),
		},
	}

	contentFetcherFactory := func(r *http.Request) FetchResultSupplier {
		return MockFetchResultSupplier{
			FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: content,
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	aggregator.ServeHTTP(resp, &http.Request{})

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
			FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: &MemoryContent{},
				Err:     nil,
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))
	aggregator.contentMergerFactory = func() ContentMerger {
		merger := NewMockContentMerger(ctrl)
		merger.EXPECT().AddContent(gomock.Any())
		merger.EXPECT().WriteHtml(gomock.Any()).Return(errors.New("an error"))
		return merger
	}

	resp := httptest.NewRecorder()
	aggregator.ServeHTTP(resp, &http.Request{})

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
			FetchResult{
				Def:     NewFetchDefinition("/foo"),
				Content: nil,
				Err:     errors.New(errorString),
			},
		}
	}
	aggregator := NewCompositionHandler(ContentFetcherFactory(contentFetcherFactory))

	resp := httptest.NewRecorder()
	aggregator.ServeHTTP(resp, &http.Request{})

	a.Equal("Bad Gateway: "+errorString+"\n", string(resp.Body.Bytes()))
	a.Equal(502, resp.Code)
}

type MockFetchResultSupplier []FetchResult

func (m MockFetchResultSupplier) WaitForResults() []FetchResult {
	return []FetchResult(m)
}
