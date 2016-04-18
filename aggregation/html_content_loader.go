package aggregation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	UiaRemove      = "uia-remove"
	UiaInclude     = "uia-include"
	UiaFragment    = "uia-fragment"
	UiaTail        = "uia-tail"
	ScriptTypeMeta = "text/uia-meta"
)

type HtmlContentLoader struct {
}

// TODO: Include Cookies and HTTP Headers from original request to the call
func (loader *HtmlContentLoader) Load(url string, timeout time.Duration) (Content, error) {

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("(http %v) on loading url %q", resp.StatusCode, url)
	}

	defer func() {
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}()
	return loader.parse(resp.Body)
}

func (loader *HtmlContentLoader) parse(in io.Reader) (Content, error) {
	z := html.NewTokenizer(in)
	c := NewMemoryContent()
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			if z.Err() == io.EOF {
				return c, nil
			}
			return nil, z.Err()
		case tt == html.StartTagToken:
			tag, _ := z.TagName()
			switch string(tag) {
			case "head":
				if err := loader.parseHead(z, c); err != nil {
					return nil, err
				}
			case "body":
				if err := loader.parseBody(z, c); err != nil {
					return nil, err
				}
			}
		}
	}
}

func (loader *HtmlContentLoader) parseBody(z *html.Tokenizer, c *MemoryContent) error {
	attrs := make([]html.Attribute, 0, 10)

forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if skipSubtreeIfUiaRemove(z, tt, string(tag), attrs) {
				continue
			}

			if a, exists := getAttr(attrs, UiaFragment); exists {
				if f, deps, err := parseFragment(z); err != nil {
					return err
				} else {
					c.body[a.Val] = f
					for _, dep := range deps {
						c.requiredContent[dep.URL] = dep
					}
				}
			}

			if _, exists := getAttr(attrs, UiaTail); exists {
				if f, deps, err := parseFragment(z); err != nil {
					return err
				} else {
					c.tail = f
					for _, dep := range deps {
						c.requiredContent[dep.URL] = dep
					}
				}
			}
		}
	}

	return nil
}

func parseFragment(z *html.Tokenizer) (f Fragment, dependencies []*FetchDefinition, err error) {
	attrs := make([]html.Attribute, 0, 10)
	depth := 0
	dependencies = make([]*FetchDefinition, 0, 0)

	buff := bytes.NewBuffer(z.Raw())
forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return nil, nil, z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if string(tag) == UiaInclude {
				fd := &FetchDefinition{}
				if url, hasUrl := getAttr(attrs, "src"); !hasUrl {
					return nil, nil, fmt.Errorf("include definition withour src %s", z.Raw())
				} else {
					fd.URL = url.Val
				}

				if timeout, hasTimeout := getAttr(attrs, "timeout"); hasTimeout {
					if timeoutInt, err := strconv.Atoi(timeout.Val); err != nil {
						return nil, nil, fmt.Errorf("error parsing timeout in %s: %s", z.Raw(), err.Error())
					} else {
						fd.Timeout = time.Millisecond * time.Duration(timeoutInt)
					}
				}

				if required, hasRequired := getAttr(attrs, "required"); hasRequired {
					if requiredBool, err := strconv.ParseBool(required.Val); err != nil {
						return nil, nil, fmt.Errorf("error parsing bool in %s: %s", z.Raw(), err.Error())
					} else {
						fd.Required = requiredBool
					}
				}

				dependencies = append(dependencies, fd)

				fmt.Fprintf(buff, "§[> %s]§", fd.URL)
				continue
			}

			if tt != html.SelfClosingTagToken {
				depth++
			}

			if skipSubtreeIfUiaRemove(z, tt, string(tag), attrs) {
				continue
			}

		case tt == html.EndTagToken:
			if depth--; depth < 0 {
				buff.Write(z.Raw())
				break forloop
			}
		}

		buff.Write(z.Raw())
	}

	return StringFragment(buff.String()), dependencies, nil
}

func (loader *HtmlContentLoader) parseHead(z *html.Tokenizer, c *MemoryContent) error {
	attrs := make([]html.Attribute, 0, 10)
	depth := 0
	headBuff := bytes.NewBuffer(nil)

forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if tt != html.SelfClosingTagToken {
				depth++
			}
			if skipSubtreeIfUiaRemove(z, tt, string(tag), attrs) {
				continue
			}

			if string(tag) == "script" && attrHasValue(attrs, "type", ScriptTypeMeta) {
				if err := parseMetaJson(z, c); err != nil {
					return err
				}
				continue
			}

		case tt == html.EndTagToken:
			if depth--; depth <= 0 {
				break forloop
			}
		}

		headBuff.Write(z.Raw())
	}

	s := headBuff.String()
	st := strings.Trim(s, " \n")
	if len(st) > 0 {
		c.head = StringFragment(st)
	}
	return nil
}

func parseMetaJson(z *html.Tokenizer, c *MemoryContent) error {
	tt := z.Next()
	if tt != html.TextToken {
		return fmt.Errorf("expected text node for meta json, but found %v, (%s)", tt.String(), z.Raw())
	}

	bytes := z.Text()
	err := json.Unmarshal(bytes, &c.meta)
	if err != nil {
		return fmt.Errorf("error while parsing json from meta json element: %v", err.Error())
	}

	tt = z.Next()
	tag, _ := z.TagName()
	if tt != html.EndTagToken || string(tag) != "script" {
		return fmt.Errorf("Tag not properly ended. Expected </script>, but found %s", z.Raw())
	}

	return nil
}

func skipSubtreeIfUiaRemove(z *html.Tokenizer, tt html.TokenType, tagName string, attrs []html.Attribute) bool {
	_, foundRemoveTag := getAttr(attrs, UiaRemove)
	if !foundRemoveTag {
		return false
	}

	if tt == html.SelfClosingTagToken {
		return true
	}

	depth := 0
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return true
		case tt == html.StartTagToken:
			depth++
		case tt == html.EndTagToken:
			depth--
			if depth < 0 {
				return true
			}
		}
	}
}

func getAttr(attrs []html.Attribute, name string) (attr html.Attribute, found bool) {
	for _, a := range attrs {
		if a.Key == name {
			return a, true
		}
	}
	return html.Attribute{}, false
}

func attrHasValue(attrs []html.Attribute, name string, value string) (found bool) {
	a, found := getAttr(attrs, name)
	return found && a.Val == value
}

func readAttributes(z *html.Tokenizer, buff []html.Attribute) []html.Attribute {
	buff = buff[:0]
	for {
		key, value, more := z.TagAttr()
		if key != nil {
			buff = append(buff, html.Attribute{Key: string(key), Val: string(value)})
		}

		if !more {
			return buff
		}
	}
}
