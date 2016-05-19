package composition

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FetchDefinition is a descriptor for fetching Content from an endpoint.
type FetchDefinition struct {
	URL      string
	Timeout  time.Duration
	Required bool
	Header   http.Header
	Method   string
	Body     io.Reader
	//ServeResponseHeaders bool
	//IsPrimary            bool
	//FallbackURL string
}

func NewFetchDefinition(url string) *FetchDefinition {
	return &FetchDefinition{
		URL:      url,
		Timeout:  10 * time.Second,
		Required: true,
		Method:   "GET",
	}
}

// NewFetchDefinitionFromRequest creates a fetch definition
// from the request path, but replaces the sheme, host and port with the provided base url
func NewFetchDefinitionFromRequest(baseUrl string, r *http.Request) *FetchDefinition {
	if strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl[:len(baseUrl)-1]
	}

	fullPath := r.URL.Path
	if fullPath == "" {
		fullPath = "/"
	}
	if r.URL.RawQuery != "" {
		fullPath += "?" + r.URL.RawQuery
	}

	return &FetchDefinition{
		URL:      baseUrl + fullPath,
		Timeout:  10 * time.Second,
		Header:   r.Header,
		Method:   r.Method,
		Body:     r.Body,
		Required: true,
	}
}

// Hash returns a unique hash for the fetch request.
// If two hashes of fetch resources are equal, they refer the same resource
// and can e.g. be taken as replacement for each other. E.g. in case of caching.
// TODO: Maybe we should exclude some headers from the hash?
func (def *FetchDefinition) Hash() string {
	hasher := md5.New()
	hasher.Write([]byte(def.URL))
	def.Header.Write(hasher)
	return hex.EncodeToString(hasher.Sum(nil))
}

// IsFetchable returns, whether the fetch definition refers to a fetchable resource
// or is a local name only.
func (def *FetchDefinition) IsFetchable() bool {
	return len(def.URL) > 0
}

type FetchResult struct {
	Def     *FetchDefinition
	Err     error
	Content Content
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
	contentLoader ContentLoader
}

// NewContentFetcher creates a ContentFetcher with an HtmlContentParser as default.
// TODO: The FetchResults should always be returned in a predictable order,
// independent of the actual response times of the fetch jobs.
func NewContentFetcher(defaultMetaJSON map[string]interface{}) *ContentFetcher {
	f := &ContentFetcher{}
	f.r.results = make([]*FetchResult, 0, 0)
	f.contentLoader = NewHttpContentLoader()
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

func (fetcher *ContentFetcher) AddFetchJob(d *FetchDefinition) {
	fetcher.AddFetchJobWithPostProcessing(d, nil)
}

// AddFetchJob addes one job to the fetcher and recursively adds the dependencies also.
func (fetcher *ContentFetcher) AddFetchJobWithPostProcessing(d *FetchDefinition, rp ResponseProcessor) {
	fetcher.r.mutex.Lock()
	defer fetcher.r.mutex.Unlock()

	hash := d.Hash()
	if fetcher.isAlreadySheduled(hash) {
		return
	}

	fetcher.activeJobs.Add(1)

	fetchResult := &FetchResult{Def: d, Hash: hash, Err: errors.New("not fetched")}
	fetcher.r.results = append(fetcher.r.results, fetchResult)

	go func() {
		defer fetcher.activeJobs.Done()

		start := time.Now()
		url, err := fetcher.expandTemplateVars(d.URL)
		if err != nil {
			log.WithField("error", err).
				WithField("fetchDefinition", d).
				Warnf("error expanding url template %v", d.URL)
			return
		}

		d.URL = url
		fetchResult.Content, fetchResult.Err = fetcher.fetch(d, rp)

		if fetchResult.Err == nil {
			log.WithField("duration", time.Since(start)).Debugf("fetched %v", d.URL)

			fetcher.addMeta(fetchResult.Content.Meta())

			for _, dependency := range fetchResult.Content.RequiredContent() {
				if dependency.IsFetchable() {
					fetcher.AddFetchJob(dependency)
				}
			}
		} else {
			log.WithField("duration", time.Since(start)).
				WithField("error", fetchResult.Err).
				WithField("fetchDefinition", d).
				Warnf("failed fetching %v", d.URL)
		}
	}()
}

func (fetcher *ContentFetcher) fetch(fd *FetchDefinition, rp ResponseProcessor) (Content, error) {
	return fetcher.contentLoader.Load(fd, rp)
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
