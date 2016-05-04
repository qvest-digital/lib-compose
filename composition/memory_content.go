package composition

type MemoryContent struct {
	url             string
	requiredContent map[string]*FetchDefinition // key ist the url
	meta            map[string]interface{}
	head            Fragment
	body            map[string]Fragment
	tail            Fragment
}

func NewMemoryContent() *MemoryContent {
	return &MemoryContent{
		requiredContent: make(map[string]*FetchDefinition),
		meta:            make(map[string]interface{}),
		body:            make(map[string]Fragment),
	}
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
