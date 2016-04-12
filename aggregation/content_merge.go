package aggregation

import (
	"io"
)

// ContentMerge is a helper type for creation of a combined html document
// out of multiple Content pages.
type ContentMerge struct {
	MetaJSON map[string]interface{}
	Head     []Fragment
	Body     map[string]Fragment
	Tail     []Fragment
}

func NewContentMerge() *ContentMerge {
	return &ContentMerge{
		MetaJSON: make(map[string]interface{}),
		Head:     make([]Fragment, 0, 0),
		Body:     make(map[string]Fragment),
		Tail:     make([]Fragment, 0, 0),
	}
}

func (cntx *ContentMerge) writeHtml(w io.Writer) {
	var executeFragment func(fragmentName string)
	executeFragment = func(fragmentName string) {
		f, exist := cntx.Body[fragmentName]
		if !exist {
			// TODO: How to handle non existing fragments!
			panic("Fragment does not exist: " + fragmentName)
		}
		f.Execute(w, cntx.MetaJSON, executeFragment)
	}

	io.WriteString(w, "<html>\n  <head>\n")
	for _, f := range cntx.Head {
		f.Execute(w, cntx.MetaJSON, executeFragment)
	}
	io.WriteString(w, "  </head>\n  <body>\n")

	executeFragment("main")

	for _, f := range cntx.Tail {
		f.Execute(w, cntx.MetaJSON, executeFragment)
	}
	io.WriteString(w, "  </body>\n</html>\n")
}

func (cntx *ContentMerge) AddContent(content Content) {
	cntx.addMeta(content.Meta())
	cntx.addHead(content.Head())
	cntx.addBody(content.Body())
	cntx.addTail(content.Tail())
}

func (cntx *ContentMerge) addMeta(data map[string]interface{}) {
	for k, v := range data {
		cntx.MetaJSON[k] = v
	}
}

func (cntx *ContentMerge) addHead(f Fragment) {
	if f != nil {
		cntx.Head = append(cntx.Head, f)
	}
}

func (cntx *ContentMerge) addBody(bodyFragmentMap map[string]Fragment) {
	for name, f := range bodyFragmentMap {
		cntx.Body[name] = f
	}
}

func (cntx *ContentMerge) addTail(f Fragment) {
	if f != nil {
		cntx.Tail = append(cntx.Tail, f)
	}
}
