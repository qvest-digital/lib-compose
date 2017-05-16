package composition

import (
	"io"
)

// StringFragment is a simple template based representation of a fragment.
type StringFragment struct {
	content string
	stylesheets []string
}

func NewStringFragment(c string) *StringFragment {
	return &StringFragment {
		content: c,
		stylesheets: []string{},
	}
}

func (f *StringFragment) Content() string {
	return f.content
}

func (f *StringFragment) Stylesheets() []string {
	return f.stylesheets
}

func (f *StringFragment) AddStylesheets(stylesheets []string) {
	f.stylesheets = append(f.stylesheets, stylesheets...)
}

func (f *StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	return executeTemplate(w, f.Content(), data, executeNestedFragment)
}

// MemorySize return the estimated size in bytes, for this object in memory
func (f *StringFragment) MemorySize() int {
	return len(f.content)
}
