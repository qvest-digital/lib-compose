package composition

import (
	"io"

	"golang.org/x/net/html"
)

// StringFragment is a simple template based representation of a fragment.
type StringFragment struct {
	content     string
	stylesheets [][]html.Attribute
}

func NewStringFragment(c string) *StringFragment {
	return &StringFragment{
		content:     c,
		stylesheets: nil,
	}
}

func (f *StringFragment) Content() string {
	return f.content
}

func (f *StringFragment) SetContent(content string) {
	f.content = content
}

func (f *StringFragment) Stylesheets() [][]html.Attribute {
	return f.stylesheets
}

func (f *StringFragment) AddStylesheets(stylesheets [][]html.Attribute) {
	f.stylesheets = append(f.stylesheets, stylesheets...)
}

func (f *StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	return executeTemplate(w, f.Content(), data, executeNestedFragment)
}

// MemorySize return the estimated size in bytes, for this object in memory
func (f *StringFragment) MemorySize() int {
	return len(f.content)
}
