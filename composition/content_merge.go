package composition

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

const (
	DefaultBufferSize = 1024 * 100
)

// ContentMerge is a helper type for creation of a combined html document
// out of multiple Content pages.
type ContentMerge struct {
	MetaJSON  map[string]interface{}
	Head      []Fragment
	BodyAttrs []Fragment
	Body      map[string]Fragment
	Tail      []Fragment
	Buffered  bool
}

// NewContentMerge creates a new buffered ContentMerge
func NewContentMerge(metaJSON map[string]interface{}) *ContentMerge {
	cntx := &ContentMerge{
		MetaJSON:  metaJSON,
		Head:      make([]Fragment, 0, 0),
		BodyAttrs: make([]Fragment, 0, 0),
		Body:      make(map[string]Fragment),
		Tail:      make([]Fragment, 0, 0),
		Buffered:  true,
	}
	return cntx
}

func (cntx *ContentMerge) WriteHtml(w io.Writer) error {
	if cntx.Buffered {
		buff := bytes.NewBuffer(make([]byte, 0, DefaultBufferSize))
		if err := cntx.WriteHtmlUnbuffered(buff); err != nil {
			return err
		}
		_, err := buff.WriteTo(w)
		return err
	} else {
		return cntx.WriteHtmlUnbuffered(w)
	}
}

func (cntx *ContentMerge) WriteHtmlUnbuffered(w io.Writer) error {
	var executeFragment func(fragmentName string) error
	executeFragment = func(fragmentName string) error {
		f, exist := cntx.Body[fragmentName]
		if !exist {
			return errors.New("Fragment does not exist: " + fragmentName)
		}
		return f.Execute(w, cntx.MetaJSON, executeFragment)
	}

	if _, err := io.WriteString(w, "<html>\n  <head>\n    "); err != nil {
		return err
	}

	for _, f := range cntx.Head {
		if err := f.Execute(w, cntx.MetaJSON, executeFragment); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "\n  </head>\n  <body"); err != nil {
		return err
	}

	for _, f := range cntx.BodyAttrs {
		if _, err := io.WriteString(w, " "); err != nil {
			return err
		}

		if err := f.Execute(w, cntx.MetaJSON, executeFragment); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, ">\n    "); err != nil {
		return err
	}

	if err := executeFragment(""); err != nil {
		return err
	}

	for _, f := range cntx.Tail {
		if err := f.Execute(w, cntx.MetaJSON, executeFragment); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, "\n  </body>\n</html>\n"); err != nil {
		return err
	}

	return nil
}

func (cntx *ContentMerge) AddContent(fetchResult *FetchResult) {
	cntx.addHead(fetchResult.Content.Head())
	cntx.addBodyAttributes(fetchResult.Content.BodyAttributes())
	cntx.addBody(fetchResult.Def.URL, fetchResult.Content.Body())
	cntx.addTail(fetchResult.Content.Tail())
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

func (cntx *ContentMerge) addBody(url string, bodyFragmentMap map[string]Fragment) {
	for name, f := range bodyFragmentMap {
		// add twice: local and full qualified name
		cntx.Body[name] = f
		cntx.Body[urlToFragmentName(url+"#"+name)] = f
	}
}

func (cntx *ContentMerge) addTail(f Fragment) {
	if f != nil {
		cntx.Tail = append(cntx.Tail, f)
	}
}

// Returns a name from a url, which has template placeholders eliminated
func urlToFragmentName(url string) string {
	url = strings.Replace(url, `§[`, `\§\[`, -1)
	url = strings.Replace(url, `]§`, `\]\§`, -1)
	return url
}
