package composition

import (
	"io"

	"golang.org/x/net/html"
)

// StringFragment is a simple template based representation of a fragment.
type StringFragment struct {
	content    string
	linkTags   [][]html.Attribute
	scriptTags []ScriptElement
}

func NewStringFragment(c string) *StringFragment {
	return &StringFragment{
		content:    c,
		linkTags:   nil,
		scriptTags: nil,
	}
}

func (f *StringFragment) Content() string {
	return f.content
}

func (f *StringFragment) SetContent(content string) {
	f.content = content
}

func (f *StringFragment) LinkTags() [][]html.Attribute {
	return f.linkTags
}

func (f *StringFragment) ScriptElements() []ScriptElement {
	return f.scriptTags
}

func (f *StringFragment) AddLinkTags(linkTags [][]html.Attribute) {
	f.linkTags = append(f.linkTags, linkTags...)
}

func (f *StringFragment) AddScriptTags(scriptTags []ScriptElement) {
	f.scriptTags = append(f.scriptTags, scriptTags...)
}

func (f *StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	return executeTemplate(w, f.Content(), data, executeNestedFragment)
}

// MemorySize return the estimated size in bytes, for this object in memory
func (f *StringFragment) MemorySize() int {
	return len(f.content)
}
