package composition

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
)

// A ContentFetcherFactory returns a configured fetch job for a request
// which can return the fetch results.
type ContentFetcherFactory func(r *http.Request) FetchResultSupplier

type CompositionHandler struct {
	contentFetcherFactory ContentFetcherFactory
	contentMergerFactory  func() ContentMerger
}

func NewCompositionHandler(contentFetcherFactory ContentFetcherFactory) *CompositionHandler {
	return &CompositionHandler{
		contentFetcherFactory: contentFetcherFactory,
		contentMergerFactory: func() ContentMerger {
			return NewContentMerge()
		},
	}
}

func (agg *CompositionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fetcher := agg.contentFetcherFactory(r)
	mergeContext := agg.contentMergerFactory()

	// fetch all contents
	results := fetcher.WaitForResults()
	for _, res := range results {
		if res.Err == nil && res.Content != nil {

			mergeContext.AddContent(res.Content)

		} else if res.Def.Required {
			log.WithField("fetchResult", res).Errorf("error loading content from: %v", res.Def.URL)
			http.Error(w, "Bad Gateway: "+res.Err.Error(), 502)
			return
		} else {
			log.WithField("fetchResult", res).Warnf("optional content not loaded: %v", res.Def.URL)
		}
	}

	// TODO: also writeHeaders

	err := mergeContext.WriteHtml(w)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), 500)
		return
	}

}
