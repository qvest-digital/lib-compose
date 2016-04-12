package aggregation

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_AggregationHandler_PositiveCase(t *testing.T) {
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
				Def:     &FetchDefinition{},
				Content: content,
			},
		}
	}
	aggregator := NewAggregationHandler(ContentFetcherFactory(contentFetcherFactory))

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
}

func Test_AggregationHandler_ErrorInMerging(t *testing.T) {
}

func Test_AggregationHandler_ErrorInFetching(t *testing.T) {
}

type MockFetchResultSupplier []FetchResult

func (m MockFetchResultSupplier) WaitForResults() []FetchResult {
	return []FetchResult(m)
}
