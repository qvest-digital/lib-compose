package composition

import (
	"errors"
	"sync"

	"github.com/tarent/lib-compose/logging"
	"strings"
	"net/http"
)

// IsFetchable returns, whether the fetch definition refers to a fetchable resource
// or is a local name only.
func (def *FetchDefinition) IsFetchable() bool {
	return len(def.URL) > 0
}

type FetchResult struct {
	Def     *FetchDefinition
	Err     error
	Content Content
	HttpStatus int
	Hash    string // the hash of the FetchDefinition
}

// ContentFetcher is a type, which can fetch a set of Content pages in parallel.
type ContentFetcher struct {
	activeJobs sync.WaitGroup
	r          struct {
		results []*FetchResult
		mutex   sync.Mutex
	}
	meta struct {
		json  map[string]interface{}
		mutex sync.Mutex
	}
	httpContentLoader ContentLoader
	fileContentLoader ContentLoader
}

// NewContentFetcher creates a ContentFetcher with an HtmlContentParser as default.
// TODO: The FetchResults should always be returned in a predictable order,
// independent of the actual response times of the fetch jobs.
func NewContentFetcher(defaultMetaJSON map[string]interface{}) *ContentFetcher {
	f := &ContentFetcher{}
	f.r.results = make([]*FetchResult, 0, 0)
	f.httpContentLoader = NewHttpContentLoader()
	f.fileContentLoader = NewFileContentLoader()
	f.meta.json = defaultMetaJSON
	if f.meta.json == nil {
		f.meta.json = make(map[string]interface{})
	}
	return f
}

// Wait blocks until all jobs are done,
// eighter sucessful or with an error result and returns the content and errors.
// Do we need to return the Results in a special order????
func (fetcher *ContentFetcher) WaitForResults() []*FetchResult {
	fetcher.activeJobs.Wait()

	fetcher.r.mutex.Lock()
	defer fetcher.r.mutex.Unlock()

	return fetcher.r.results
}

// AddFetchJob addes one job to the fetcher and recursively adds the dependencies also.
func (fetcher *ContentFetcher) AddFetchJob(d *FetchDefinition) {
	fetcher.r.mutex.Lock()
	defer fetcher.r.mutex.Unlock()

	hash := d.Hash()
	if fetcher.isAlreadySheduled(hash) {
		return
	}

	fetcher.activeJobs.Add(1)

	fetchResult := &FetchResult{Def: d, Hash: hash, Err: errors.New("not fetched"), HttpStatus:http.StatusBadGateway}
	fetcher.r.results = append(fetcher.r.results, fetchResult)

	go func() {
		defer fetcher.activeJobs.Done()

		url, err := fetcher.expandTemplateVars(d.URL)
		if err != nil {
			logging.Logger.
				WithField("fetchDefinition", d).
				WithError(err).
				Warnf("error expanding url template %v", d.URL)
			return
		}

		// create a copy of the fetch definition, to because we do not
		// want to override the original URL with expanded values
		definitionCopy := *d
		definitionCopy.URL = url
		fetchResult.Content, fetchResult.Err, fetchResult.HttpStatus = fetcher.fetch(&definitionCopy)

		if fetchResult.Err == nil {
			fetcher.addMeta(fetchResult.Content.Meta())
			for _, dependency := range fetchResult.Content.RequiredContent() {
				if dependency.IsFetchable() {
					fetcher.AddFetchJob(dependency)
				}
			}
		} else {
			logging.Logger.WithError(fetchResult.Err).
				WithField("fetchDefinition", d).
				WithField("correlation_id", logging.GetCorrelationId(definitionCopy.Header)).
				Errorf("failed fetching %v", d.URL)
		}
	}()
}

func (fetcher *ContentFetcher) fetch(fd *FetchDefinition) (Content, error, int) {
	if strings.HasPrefix(fd.URL, FileURLPrefix) {
		return fetcher.fileContentLoader.Load(fd)
	}
	return fetcher.httpContentLoader.Load(fd)
}

// isAlreadySheduled checks, if there is already a job for a FetchDefinition, or it is already fetched.
// The method has to be called in a locked mutex block.
func (fetcher *ContentFetcher) isAlreadySheduled(fetchDefinitionHash string) bool {
	for _, fetchResult := range fetcher.r.results {
		if fetchDefinitionHash == fetchResult.Hash {
			return true
		}
	}
	return false
}

func (fetcher *ContentFetcher) MetaJSON() map[string]interface{} {
	fetcher.meta.mutex.Lock()
	defer fetcher.meta.mutex.Unlock()
	return fetcher.meta.json
}

func (fetcher *ContentFetcher) expandTemplateVars(template string) (string, error) {
	fetcher.meta.mutex.Lock()
	defer fetcher.meta.mutex.Unlock()
	return expandTemplateVars(template, fetcher.meta.json)
}

func (fetcher *ContentFetcher) addMeta(data map[string]interface{}) {
	fetcher.meta.mutex.Lock()
	defer fetcher.meta.mutex.Unlock()
	for k, v := range data {
		fetcher.meta.json[k] = v
	}
}
