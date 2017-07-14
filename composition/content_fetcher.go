package composition

import (
	"errors"
	"github.com/tarent/lib-compose/logging"
	"sort"
	"sync"
)

type FetchResult struct {
	Def     *FetchDefinition
	Err     error
	Content Content
	Hash    string // the hash of the FetchDefinition
}

//Provide implementation for sorting FetchResults by priority with sort.Sort
type FetchResults []*FetchResult

func (fr FetchResults) Len() int {
	return len(fr)
}
func (fr FetchResults) Swap(i, j int) {
	fr[i], fr[j] = fr[j], fr[i]
}
func (fr FetchResults) Less(i, j int) bool {
	return fr[i].Def.Priority < fr[j].Def.Priority
}

// FetchDefinitionFactory should return a fetch definition for the given name and parameters.
// This factory method can be used to supply lazy loaded fetch jobs.
// The FetchDefinition returned has to have the same name as the supplied name parameter.
// If no fetch definition for the supplied name can be provided by the factory, existing=false is returned, otherwise existing=true.
type FetchDefinitionFactory func(name string, params Params) (fd *FetchDefinition, existing bool, err error)

// ContentFetcher is a type, which can fetch a set of Content pages in parallel.
type ContentFetcher struct {
	activeJobs sync.WaitGroup
	r          struct {
		sheduledFetchDefinitionNames map[string]string
		results                      []*FetchResult
		mutex                        sync.Mutex
	}
	meta struct {
		json  map[string]interface{}
		mutex sync.Mutex
	}
	lazyFdFactory FetchDefinitionFactory
	Loader        ContentLoader
}

// NewContentFetcher creates a ContentFetcher with an HtmlContentParser as default.
// TODO: The FetchResults should always be returned in a predictable order,
// independent of the actual response times of the fetch jobs.
func NewContentFetcher(defaultMetaJSON map[string]interface{}, collectLinks bool, collectScripts bool) *ContentFetcher {
	f := &ContentFetcher{}
	f.r.results = make([]*FetchResult, 0, 0)
	f.r.sheduledFetchDefinitionNames = make(map[string]string)
	f.Loader = NewHttpContentLoader(collectLinks, collectScripts)
	f.meta.json = defaultMetaJSON
	if f.meta.json == nil {
		f.meta.json = make(map[string]interface{})
	}
	f.lazyFdFactory = func(name string, params Params) (fd *FetchDefinition, exist bool, err error) {
		return nil, false, nil
	}
	return f
}

// SetFetchDefinitionFactory supplies a factory for lazy evaluated fetch jobs,
// which will only be loaded if a fragment refrences them.
// Seting the factory of optional, but if used, has to be done before adding Jobs by AddFetchJob.
func (fetcher *ContentFetcher) SetFetchDefinitionFactory(factory FetchDefinitionFactory) {
	fetcher.lazyFdFactory = factory
}

// Wait blocks until all jobs are done,
// either successful or with an error result and returns the content and errors.
// Do we need to return the Results in a special order????
func (fetcher *ContentFetcher) WaitForResults() []*FetchResult {
	fetcher.activeJobs.Wait()

	fetcher.r.mutex.Lock()
	defer fetcher.r.mutex.Unlock()

	results := fetcher.r.results

	// To keep initial order if no priority settings are given, do a check before for sorting.
	if hasPrioritySetting(results) {
		sort.Sort(FetchResults(results))
	}

	return results
}

//func (fetcher *ContentFetcher) AddFetchDefinitionFactory(name string, func(params map[string]string) *FetchDefinition) {

// AddFetchJob adds one job to the fetcher and recursively adds the dependencies also.
func (fetcher *ContentFetcher) AddFetchJob(d *FetchDefinition) {
	fetcher.r.mutex.Lock()
	defer fetcher.r.mutex.Unlock()

	hash := d.Hash()
	if fetcher.isAlreadyScheduled(hash) {
		return
	}

	fetcher.activeJobs.Add(1)
	fetchResult := &FetchResult{Def: d, Hash: hash, Err: errors.New("not fetched")}
	fetcher.r.results = append(fetcher.r.results, fetchResult)
	fetcher.r.sheduledFetchDefinitionNames[d.Name] = d.Name

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

		// Create a copy of the fetch definition, to because we do not
		// want to override the original URL with expanded values.
		definitionCopy := *d
		definitionCopy.URL = url
		fetchResult.Content, fetchResult.Err = fetcher.Loader.Load(&definitionCopy)

		if fetchResult.Err == nil {
			fetcher.addMeta(fetchResult.Content.Meta())
			fetcher.addDependentFetchJobs(fetchResult.Content)
		} else {
			// 404 Error already become logged in logger.go
			if fetchResult.Content == nil || fetchResult.Content.HttpStatusCode() != 404 {
				logging.Logger.WithError(fetchResult.Err).
					WithField("fetchDefinition", d).
					WithField("correlation_id", logging.GetCorrelationId(definitionCopy.Header)).
					Errorf("failed fetching %v", d.URL)
			}
		}
	}()
}

func (fetcher *ContentFetcher) addDependentFetchJobs(content Content) {
	for _, fetch := range content.RequiredContent() {
		fetcher.AddFetchJob(fetch)
	}
	for dependencyName, params := range content.Dependencies() {
		fetcher.r.mutex.Lock()
		_, alreadySheduled := fetcher.r.sheduledFetchDefinitionNames[dependencyName]
		fetcher.r.mutex.Unlock()
		if !alreadySheduled {
			lazyFd, existing, err := fetcher.lazyFdFactory(dependencyName, params)
			if err != nil {
				logging.Logger.WithError(err).
					WithField("dependencyName", dependencyName).
					WithField("params", params).
					Errorf("failed optaining a fetch definition for dependency %v", dependencyName)
			}
			if err == nil && existing {
				fetcher.AddFetchJob(lazyFd)
			}
			// error handling: In the case, the fd could not be loaded, we will do
			// the error handling in the merging process.
		}
	}
}

func (fetcher *ContentFetcher) Empty() bool {
	fetcher.r.mutex.Lock()
	defer fetcher.r.mutex.Unlock()
	return len(fetcher.r.results) == 0
}

// isAlreadyScheduled checks if there is already a job for a FetchDefinition, or it is already fetched.
// The method has to be called in a locked mutex block.
func (fetcher *ContentFetcher) isAlreadyScheduled(fetchDefinitionHash string) bool {
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
