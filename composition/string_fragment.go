package composition

import (
	"io"
)

// StringFragment is a simple template based representation of a fragment.
// On Execute(), the following replacements will be done:
// §[ aVariable ] inserts a variable from the data map
// §[> fragment ] executed a nexted fragment by executeNestedFragment()
type StringFragment string

func (f StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	w.Write([]byte(f))
	return nil
}

func (f StringFragment) String() string {
	return string(f)
}
