package aggregation

import (
	"io"
	"time"
)

type Fragment interface {
	Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string)) error
}

type ContentLoader interface {
	// Load synchronously loads a content.
	// The loader has to ensure to return the call at withing the supplied timeout.
	Load(url string, timeout time.Duration) (Content, error)
}

type Content interface {

	// RequiredContent returns a list of Content Elements to load
	RequiredContent() []*FetchDefinition

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
}
