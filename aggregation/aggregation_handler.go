package aggregation

import (
	"net/http"
)

// A ContentFetcherFactory returns a configured fetch job for a request
// which can return the fetch results.
type ContentFetcherFactory func(r *http.Request) FetchResultSupplier

type AggregationHandler struct {
	contentFetcherFactory ContentFetcherFactory
	contentMergerFactory  func() ContentMerger
}

func NewAggregationHandler(contentFetcherFactory ContentFetcherFactory) *AggregationHandler {
	return &AggregationHandler{
		contentFetcherFactory: contentFetcherFactory,
		contentMergerFactory: func() ContentMerger {
			return NewContentMerge()
		},
	}
}

func (agg *AggregationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fetcher := agg.contentFetcherFactory(r)
	mergeContext := agg.contentMergerFactory()

	// fetch all contents
	results := fetcher.WaitForResults()
	for _, res := range results {
		if res.Err != nil && res.Def.Required {
			http.Error(w, "Bad Gateway: "+res.Err.Error(), 502)
			return
		}

		mergeContext.AddContent(res.Content)
	}

	// TODO: also writeHeaders

	err := mergeContext.WriteHtml(w)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), 500)
		return
	}

}
