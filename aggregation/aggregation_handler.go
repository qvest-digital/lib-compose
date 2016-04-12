package aggregation

import (
	"net/http"
)

// A ContentFetcherFactory returns a configured fetch job for a request
// which can return the fetch results.
type ContentFetcherFactory func(r *http.Request) FetchResultSupplier

type AggregationHandler struct {
	contentFetcherFactory ContentFetcherFactory
	contentMergeFactory   func() *ContentMerge
}

func NewAggregationHandler(contentFetcherFactory ContentFetcherFactory) *AggregationHandler {
	return &AggregationHandler{
		contentFetcherFactory: contentFetcherFactory,
		contentMergeFactory: func() *ContentMerge {
			return NewContentMerge()
		},
	}
}

func (agg *AggregationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fetcher := agg.contentFetcherFactory(r)
	mergeContext := agg.contentMergeFactory()

	// fetch all contents
	results := fetcher.WaitForResults()
	for _, res := range results {
		if res.Err != nil && res.Def.Required {
			http.Error(w, res.Err.Error(), 500)
			return
		}

		mergeContext.AddContent(res.Content)
	}

	// TODO: also writeHeaders
	mergeContext.writeHtml(w)
}
