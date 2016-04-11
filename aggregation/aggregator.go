package aggregation

import (
	"io"
	"net/http"
)

type Aggregator struct {
	// fetcherForRequest, returns a configured fetcher for a request
	fetcherForRequest func(r *http.Request) *ContentFetcher
}

func (agg *Aggregator) doAggregation(w http.ResponseWriter, r *http.Request) {

	fetcher := agg.fetcherForRequest(r)
	mergeContext := newMergeContext()

	// fetch all contents
	results := fetcher.WaitForResults()
	for _, res := range results {
		if res.err != nil && res.def.Required {
			http.Error(w, res.err.Error(), 500)
			return
		}

		mergeContext.AddContent(res.content)
	}

	agg.writeHeaders(w, mergeContext)
	agg.writeHtml(w, mergeContext)
}

func (agg *Aggregator) writeHeaders(w http.ResponseWriter, cntx *mergeContext) {
	// TODO
}

func (agg *Aggregator) writeHtml(w http.ResponseWriter, cntx *mergeContext) {
	var executeFragment func(fragmentName string)
	executeFragment = func(fragmentName string) {
		f, exist := cntx.Body[fragmentName]
		if !exist {
			// TODO: How to handle non existing fragments!
			panic("Fragment does not exist: " + fragmentName)
		}
		f.Execute(w, cntx.MetaJSON, executeFragment)
	}

	io.WriteString(w, "<html>\n  <head>")
	for _, f := range cntx.Head {
		f.Execute(w, cntx.MetaJSON, executeFragment)
	}
	io.WriteString(w, "  </head>\n   <body>")

	executeFragment("main")

	for _, f := range cntx.Tail {
		f.Execute(w, cntx.MetaJSON, executeFragment)
	}
	io.WriteString(w, "  </body>\n</html>")
}

type mergeContext struct {
	MetaJSON map[string]interface{}
	Head     []Fragment
	Body     map[string]Fragment
	Tail     []Fragment
}

func newMergeContext() *mergeContext {
	return &mergeContext{
		MetaJSON: make(map[string]interface{}),
		Head:     make([]Fragment, 0, 5),
		Body:     make(map[string]Fragment),
		Tail:     make([]Fragment, 0, 5),
	}
}

func (cntx *mergeContext) AddContent(content Content) {
	cntx.addMeta(content.Meta())
	cntx.addHead(content.Head())
	cntx.addBody(content.Body())
	cntx.addTail(content.Tail())
}

func (cntx *mergeContext) addMeta(data map[string]interface{}) {
	for k, v := range data {
		cntx.MetaJSON[k] = v
	}
}

func (cntx *mergeContext) addHead(f Fragment) {
	cntx.Head = append(cntx.Head, f)
}

func (cntx *mergeContext) addBody(bodyFragmentMap map[string]Fragment) {
	for name, f := range bodyFragmentMap {
		cntx.Body[name] = f
	}
}

func (cntx *mergeContext) addTail(f Fragment) {
	cntx.Tail = append(cntx.Tail, f)
}
