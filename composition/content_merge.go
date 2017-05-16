package composition

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

const (
	LayoutFragmentName = "layout"
	FragmentSeparater  = "#"
	DefaultBufferSize  = 1024 * 100
)

// ContentMerge is a helper type for creation of a combined html document
// out of multiple Content pages.
type ContentMerge struct {
	MetaJSON  map[string]interface{}
	Head      []Fragment
	BodyAttrs []Fragment

	// Aggregator for the Body Fragments of the results.
	// Each fragment is insertes twice with full name and local name,
	// The full name only ends with a FragmentSeparater ('#'), if the local name is not empty
	// and the local name is always prefixed with FragmentSeparater ('#').
	Body map[string]Fragment

	// Aggregator for the Tail Fragments of the results.
	Tail     []Fragment
	Buffered bool

	// merge priorities for the content objects
	// no entry means priority == 0
	priorities map[Content]int
}

// NewContentMerge creates a new buffered ContentMerge
func NewContentMerge(metaJSON map[string]interface{}) *ContentMerge {
	cntx := &ContentMerge{
		MetaJSON:   metaJSON,
		Head:       make([]Fragment, 0, 0),
		BodyAttrs:  make([]Fragment, 0, 0),
		Body:       make(map[string]Fragment),
		Tail:       make([]Fragment, 0, 0),
		Buffered:   true,
		priorities: make(map[Content]int),
	}
	return cntx
}

func (cntx *ContentMerge) GetHtml() ([]byte, error) {
	if len(cntx.priorities) > 0 {
		cntx.processMetaPriorityParsing()
	}
	w := bytes.NewBuffer(make([]byte, 0, DefaultBufferSize))

	var executeFragment func(fragmentName string) error
	executeFragment = func(fragmentName string) error {
		f, exist := cntx.GetBodyFragmentByName(fragmentName)
		if !exist {
			missingFragmentString := generateMissingFragmentString(cntx.Body, fragmentName)
			return errors.New(missingFragmentString)
		}
		return f.Execute(w, cntx.MetaJSON, executeFragment)
	}

	io.WriteString(w, "<!DOCTYPE html>\n<html>\n  <head>\n    ")

	for _, f := range cntx.Head {
		if err := f.Execute(w, cntx.MetaJSON, executeFragment); err != nil {
			return nil, err
		}
	}
	io.WriteString(w, "\n  </head>\n  <body")

	for _, f := range cntx.BodyAttrs {
		io.WriteString(w, " ")

		if err := f.Execute(w, cntx.MetaJSON, executeFragment); err != nil {
			return nil, err
		}
	}

	io.WriteString(w, ">\n    ")

	startFragmentName := ""
	if _, exist := cntx.GetBodyFragmentByName(LayoutFragmentName); exist {
		startFragmentName = LayoutFragmentName
	}

	if err := executeFragment(startFragmentName); err != nil {
		return nil, err
	}

	for _, f := range cntx.Tail {
		if err := f.Execute(w, cntx.MetaJSON, executeFragment); err != nil {
			return nil, err
		}
	}

	io.WriteString(w, "\n  </body>\n</html>\n")

	return w.Bytes(), nil
}

// GetBodyFragmentByName returns a fragment by ists name.
// If the name does not contain a FragmentSeparater ('#'), and no such fragment is found.
// also a lookup for '#name' is done, to check, if there is a local name matching.
// The bool return value indicates, if the fragment was found.
func (cntx *ContentMerge) GetBodyFragmentByName(name string) (Fragment, bool) {
	f, found := cntx.Body[name]

	// Normalize: e.g. main# -> main
	if !found && strings.HasSuffix(name, FragmentSeparater) {
		f, found = cntx.Body[name[0:len(name)-1]]
	}

	// search also for local fragment if nothing else found
	if !found && !strings.Contains(name, FragmentSeparater) {
		f, found = cntx.Body[FragmentSeparater+name]
	}

	return f, found
}

func (cntx *ContentMerge) AddContent(c Content, priority int) {
	cntx.addHead(c.Head())
	cntx.addBodyAttributes(c.BodyAttributes())
	cntx.addBody(c)
	cntx.addTail(c.Tail())
	if priority > 0 {
		cntx.priorities[c] = priority
	}
}

func (cntx *ContentMerge) addHead(f Fragment) {
	if f != nil {
		cntx.Head = append(cntx.Head, f)
	}
}

func (cntx *ContentMerge) addBodyAttributes(f Fragment) {
	if f != nil {
		cntx.BodyAttrs = append(cntx.BodyAttrs, f)
	}
}

func (cntx *ContentMerge) addBody(c Content) {

	for localName, f := range c.Body() {
		// add twice: local and full qualified name
		cntx.Body[FragmentSeparater+localName] = f
		fqn := c.Name()
		if localName != "" {
			fqn += FragmentSeparater + localName
		}
		cntx.Body[fqn] = f
	}
}

func (cntx *ContentMerge) addTail(f Fragment) {
	if f != nil {
		cntx.Tail = append(cntx.Tail, f)
	}
}

// Generates String for the missing Fragment error message. It adds all existing fragments from the body
func generateMissingFragmentString(body map[string]Fragment, fragmentName string) string {
	text := "Fragment does not exist: " + fragmentName + ". Existing fragments: "
	index := 0
	for k, _ := range body {
		if index == 0 {
			text += `"` + k + `"`
		} else {
			text += `, "` + k + `"`
		}
		index++
	}
	return text
}

// Processes all heads to remove duplicate meta and title tags, respecting the priority of head fragments
func (cntx *ContentMerge) processMetaPriorityParsing() {
	headPropertyMap := make(map[string]string)

	for i := len(cntx.Head) - 1; i >= 0; i-- {
		var currentHead interface{} = cntx.Head[i]
		if currentHead != nil {
			currentStringFragment := currentHead.(*StringFragment)
			ParseHeadFragment(currentStringFragment, headPropertyMap)
			cntx.Head[i] = currentStringFragment
		}
	}
}
