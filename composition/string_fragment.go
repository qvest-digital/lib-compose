package composition

import (
	"io"
)

// StringFragment is a simple template based representation of a fragment.
type StringFragment string

func (f StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	return executeTemplate(w, string(f), data, executeNestedFragment)
}
