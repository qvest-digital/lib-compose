package composition

import (
	"io"
	"net/http"
)

type MemoryContent struct {
	url             string
	requiredContent map[string]*FetchDefinition // key ist the url
	meta            map[string]interface{}
	head            Fragment
	body            map[string]Fragment
	tail            Fragment
	bodyAttributes  Fragment
	reader          io.ReadCloser
	httpHeader      http.Header
	httpStatusCode  int
}

func NewMemoryContent() *MemoryContent {
	return &MemoryContent{
		requiredContent: make(map[string]*FetchDefinition),
		meta:            make(map[string]interface{}),
		body:            make(map[string]Fragment),
	}
}

func (c *MemoryContent) MemorySize() int {
	// We estimate the size for caching, here
	// so a rougth esitmation is enough
	i := len(c.meta)*20 + len(c.httpHeader)*20

	if c.head != nil {
		i += c.head.MemorySize()
	}
	if c.tail != nil {
		i += c.tail.MemorySize()
	}
	for _, f := range c.body {
		i += f.MemorySize()
	}
	return i
}

func (c *MemoryContent) URL() string {
	return c.url
}

func (c *MemoryContent) RequiredContent() []*FetchDefinition {
	deps := make([]*FetchDefinition, 0, len(c.requiredContent))
	for _, dep := range c.requiredContent {
		deps = append(deps, dep)
	}
	return deps
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

func (c *MemoryContent) BodyAttributes() Fragment {
	return c.bodyAttributes
}

func (c *MemoryContent) Reader() io.ReadCloser {
	return c.reader
}

func (c *MemoryContent) SetReader(r io.ReadCloser) {
	c.reader = r
}

func (c *MemoryContent) HttpHeader() http.Header {
	return c.httpHeader
}

func (c *MemoryContent) HttpStatusCode() int {
	return c.httpStatusCode
}
