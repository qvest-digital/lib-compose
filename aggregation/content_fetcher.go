package aggregation

import (
	"net/http"
	"sync"
	"time"
)

type FetchDefinition struct {
	URL            string
	Timeout        time.Duration
	RequestHeaders http.Header
	Required       bool
	//ServeResponseHeaders bool
	//IsPrimary            bool
	//FallbackURL string
}

type FetchResult struct {
	def     *FetchDefinition
	err     error
	content Content
}

func NewFetchDefinition(url string) *FetchDefinition {
	return &FetchDefinition{
		URL:            url,
		Timeout:        time.Second,
		RequestHeaders: nil,
		Required:       true,
	}
}

type ContentFetcher struct {
	allDone sync.WaitGroup
	r       struct {
		results []FetchResult
		mutex   sync.Mutex
	}
}

func NewContentFetcher() *ContentFetcher {
	f := &ContentFetcher{}
	f.r.results = make([]FetchResult, 0, 0)
	return f
}

// Wait blocks until all jobs are done,
// eighter sucessful or with an error result and returns the content and errors.
// Do we need to return the Results in a special order????
func (fetcher *ContentFetcher) WaitForResults() []FetchResult {
	fetcher.allDone.Wait()
	fetcher.r.mutex.Lock()

	defer fetcher.r.mutex.Unlock()
	return fetcher.r.results
}

// AddFetchJob addes one job to the fetcher and recursively adds the dependencies also.
func (fetcher *ContentFetcher) AddFetchJob(d *FetchDefinition) *ContentFetcher {
	fetcher.allDone.Add(1)
	go func() {
		defer fetcher.allDone.Done()
		loader := &HtmlContentLoader{}

		result := FetchResult{def: d}
		result.content, result.err = loader.Load(d.URL, d.Timeout)

		fetcher.r.mutex.Lock()
		defer fetcher.r.mutex.Unlock()

		for _, dependency := range result.content.RequiredContent() {
			fetcher.AddFetchJob(dependency)
		}

		fetcher.r.results = append(fetcher.r.results, result)
	}()
	return fetcher
}
