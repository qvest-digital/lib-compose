package composition

//go:generate go get github.com/golang/mock/mockgen
//go:generate mockgen -self_package composition -package composition -destination interface_mocks_test.go github.com/tarent/lib-compose/composition Fragment,ContentLoader,Content,ContentMerger,ContentParser,ResponseProcessor,Cache
//go:generate sed -ie "s/composition .github.com\\/tarent\\/lib-compose\\/composition.//g;s/composition\\.//g" interface_mocks_test.go
//go:generate sh ../scripts/mockgen.sh

import (
	"io"
	"net/http"

	"golang.org/x/net/html"
)

type Fragment interface {
	Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error

	// MemorySize return the estimated size in bytes, for this object in memory
	MemorySize() int

	// Return the content of this fragment
	Content() string

	// Return the list of link tags and script elements used in this fragment
	LinkTags() [][]html.Attribute
	ScriptElements() []ScriptElement
}

type ContentLoader interface {
	// Load synchronously loads a content.
	// The loader has to ensure to return the call withing the supplied timeout.
	Load(fd *FetchDefinition) (content Content, err error)
}

type ContentParser interface {
	// Parse parses the input stream into a Content Object
	Parse(*MemoryContent, io.Reader) error
}

type FetchResultSupplier interface {
	// WaitForResults returns all results of a fetch job in a blocking manger.
	WaitForResults() []*FetchResult

	// MetaJSON returns the composed meta JSON object
	MetaJSON() map[string]interface{}

	// True, if no fetch jobs were added
	Empty() bool
}

type CacheStrategy interface {
	Hash(method string, url string, requestHeader http.Header) string
	IsCacheable(method string, url string, statusCode int, requestHeader http.Header, responseHeader http.Header) bool
}

// Params is a value type for a parameter map
type Params map[string]string

// Vontent is the abstration over includable data.
// Content may be parsed of it may contain a stream represented by a non nil Reader(), not both.
type Content interface {

	// The Name of the content, as given in the fetch definition
	Name() string

	// RequiredContent returns a list of Content Elements to load
	RequiredContent() []*FetchDefinition

	// Dependencies returns list of referenced content element names.
	// The list only contains the base names of the includes e.g. 'foo' for '<uic-include src="foo#bar"/>'
	Dependencies() map[string]Params

	// Meta returns a data structure to add to the global
	// data context.
	Meta() map[string]interface{}

	// Head returns a partial which should be
	// inserted to the html head
	Head() Fragment

	// Body returns a map of partials,
	// the named body partials, where the keys are partial names.
	Body() map[string]Fragment

	// Tail returns a partial which should be inserted at the end of the page.
	// e.g. a script to load after rendering.
	Tail() Fragment

	// The attributes for the body element
	BodyAttributes() Fragment

	// Reader returns the stream with the content, of any.
	// If Reader() == nil, no stream is available an it contains parsed data, only.
	Reader() io.ReadCloser

	// HttpHeader() returns the http headers of the fetch job
	HttpHeader() http.Header

	// HttpStatusCode() returns the http statuc code of the fetch job
	HttpStatusCode() int

	// MemorySize return the estimated size in bytes, for this object in memory
	MemorySize() int
}

type ContentMerger interface {
	// Add content to the merger
	AddContent(c Content, priority int)

	// Return the html as byte array
	GetHtml() ([]byte, error)

	// Set the stratgy for stylesheet deduplication
	SetDeduplicationStrategy(strategy DeduplicationStrategy)
}

type ResponseProcessor interface {
	// Process html from responsebody before composition is triggered
	// May create a new Reader inside the ResponseBody
	Process(*http.Response, string) error
}

type ErrorHandler interface {
	// handle http request errors
	Handle(err error, status int, w http.ResponseWriter, r *http.Request)
}

type Cache interface {
	Get(hash string) (cacheObject interface{}, found bool)
	Set(hash string, label string, memorySize int, cacheObject interface{})
	Invalidate()
	PurgeEntries(keys []string)
}

type DeduplicationStrategy interface {
	Deduplicate(linkTags [][]html.Attribute) [][]html.Attribute
	DeduplicateElements(scriptElements []ScriptElement) []ScriptElement
}
