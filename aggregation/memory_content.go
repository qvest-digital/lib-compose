package aggregation

import (
	"io"
)

type StringFragment string

func (f StringFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string)) error {
	w.Write([]byte(f))
	return nil
}

type MemoryContent struct {
	requiredContent []*FetchDefinition
	meta            map[string]interface{}
	head            Fragment
	body            map[string]Fragment
	tail            Fragment
}

func NewMemoryContent() *MemoryContent {
	return &MemoryContent{
		requiredContent: make([]*FetchDefinition, 0, 0),
		meta:            make(map[string]interface{}),
		head:            nil,
		body:            make(map[string]Fragment),
		tail:            nil,
	}
}

func (c *MemoryContent) RequiredContent() []*FetchDefinition {
	return c.requiredContent
}

func (c *MemoryContent) Meta() map[string]interface{} {
	return c.meta
}

func (c *MemoryContent) Head() Fragment {
	return c.head
}

func (c *MemoryContent) Body() map[string]Fragment {
	return c.body
}

func (c *MemoryContent) Tail() Fragment {
	return c.tail
}
