package composition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	UicRemove       = "uic-remove"
	UicInclude      = "uic-include"
	UicFetch        = "uic-fetch"
	UicFragment     = "uic-fragment"
	UicTail         = "uic-tail"
	ScriptTypeMeta  = "text/uic-meta"
	ParamAttrPrefix = "param-"
)

type HtmlContentParser struct {
}

func (parser *HtmlContentParser) Parse(c *MemoryContent, in io.Reader) error {
	z := html.NewTokenizer(in)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		case tt == html.StartTagToken:
			tag, _ := z.TagName()
			switch string(tag) {
			case "head":
				if err := parser.parseHead(z, c); err != nil {
					return err
				}
			case "body":
				if err := parser.parseBody(z, c); err != nil {
					return err
				}
			}
		}
	}
}

func (parser *HtmlContentParser) parseHead(z *html.Tokenizer, c *MemoryContent) error {
	attrs := make([]html.Attribute, 0, 10)
	headBuff := bytes.NewBuffer(nil)

forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		raw := byteCopy(z.Raw()) // create a copy here, because readAttributes modifies z.Raw, if attributes contain an &
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if skipSubtreeIfUicRemove(z, tt, string(tag), attrs) {
				continue
			}
			if string(tag) == "script" && attrHasValue(attrs, "type", ScriptTypeMeta) {
				if err := parseMetaJson(z, c); err != nil {
					return err
				}
				continue
			}
		case tt == html.EndTagToken:
			if string(tag) == "head" {
				break forloop
			}
		}
		headBuff.Write(raw)
	}

	s := headBuff.String()
	st := strings.Trim(s, " \n")
	if len(st) > 0 {
		c.head = StringFragment(st)
	}
	return nil
}

func (parser *HtmlContentParser) parseBody(z *html.Tokenizer, c *MemoryContent) error {
	attrs := make([]html.Attribute, 0, 10)
	bodyBuff := bytes.NewBuffer(nil)

	attrs = readAttributes(z, attrs)
	if len(attrs) > 0 {
		c.bodyAttributes = StringFragment(joinAttrs(attrs))
	}

forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		raw := byteCopy(z.Raw()) // create a copy here, because readAttributes modifies z.Raw, if attributes contain an &
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if skipSubtreeIfUicRemove(z, tt, string(tag), attrs) {
				continue
			}
			if string(tag) == UicFragment {
				if f, deps, err := parseFragment(z); err != nil {
					return err
				} else {
					c.body[getFragmentName(attrs)] = f
					for depName, depParams := range deps {
						c.dependencies[depName] = depParams
					}
				}
				continue
			}
			if string(tag) == UicTail {
				if f, deps, err := parseFragment(z); err != nil {
					return err
				} else {
					c.tail = f
					for depName, depParams := range deps {
						c.dependencies[depName] = depParams
					}
				}
				continue
			}
			if string(tag) == UicFetch {
				if fd, err := getFetch(z, attrs); err != nil {
					return err
				} else {
					c.requiredContent[fd.URL] = fd
					continue
				}
			}
			if string(tag) == UicInclude {
				if replaceTextStart, replaceTextEnd, dependencyName, dependencyParams, err := getInclude(z, attrs); err != nil {
					return err
				} else {
					c.dependencies[dependencyName] = dependencyParams
					bodyBuff.WriteString(replaceTextStart)
					// Enhancement: WriteOut sub tree, to allow alternative content
					//              for optional includes.
					bodyBuff.WriteString(replaceTextEnd)
					continue
				}
			}

		case tt == html.EndTagToken:
			if string(tag) == "body" {
				break forloop
			}
		}
		bodyBuff.Write(raw)
	}

	s := bodyBuff.String()
	if _, defaultFragmentExists := c.body[""]; !defaultFragmentExists {
		if st := strings.Trim(s, " \n"); len(st) > 0 {
			c.body[""] = StringFragment(st)
		}
	}

	return nil
}

func parseFragment(z *html.Tokenizer) (f Fragment, dependencies map[string]Params, err error) {
	attrs := make([]html.Attribute, 0, 10)
	dependencies = make(map[string]Params)

	buff := bytes.NewBuffer(nil)
forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		raw := byteCopy(z.Raw()) // create a copy here, because readAttributes modifies z.Raw, if attributes contain an &
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return nil, nil, z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if string(tag) == UicInclude {
				if replaceTextStart, replaceTextEnd, dependencyName, dependencyParams, err := getInclude(z, attrs); err != nil {
					return nil, nil, err
				} else {
					dependencies[dependencyName] = dependencyParams
					fmt.Fprintf(buff, replaceTextStart)
					// Enhancement: WriteOut sub tree, to allow alternative content
					//              for optional includes.
					fmt.Fprintf(buff, replaceTextEnd)
					continue
				}
			}

			if skipSubtreeIfUicRemove(z, tt, string(tag), attrs) {
				continue
			}

		case tt == html.EndTagToken:
			if string(tag) == UicFragment || string(tag) == UicTail {
				break forloop
			}
		}
		buff.Write(raw)
	}

	return StringFragment(buff.String()), dependencies, nil
}

func getInclude(z *html.Tokenizer, attrs []html.Attribute) (startMarker, endMarker, dependencyName string, dependencyParams Params, error error) {
	var srcString string
	if url, hasUrl := getAttr(attrs, "src"); !hasUrl {
		return "", "", "", nil, fmt.Errorf("include definition without src %s", z.Raw())
	} else {
		srcString = strings.TrimSpace(url.Val)
		if strings.HasPrefix(srcString, "#") {
			srcString = srcString[1:]
		}
		dependencyName = strings.Split(srcString, "#")[0]
	}

	dependencyParams = Params{}
	for _, a := range attrs {
		if strings.HasPrefix(a.Key, ParamAttrPrefix) {
			key := a.Key[len(ParamAttrPrefix):]
			dependencyParams[key] = a.Val
		}
	}

	required := false
	if r, hasRequired := getAttr(attrs, "required"); hasRequired {
		if requiredBool, err := strconv.ParseBool(r.Val); err != nil {
			return "", "", "", nil, fmt.Errorf("error parsing bool in %s: %s", z.Raw(), err.Error())
		} else {
			required = requiredBool
		}
	}

	if required {
		return fmt.Sprintf("§[> %s]§", srcString), "", dependencyName, dependencyParams, nil
	} else {
		return fmt.Sprintf("§[#> %s]§", srcString), fmt.Sprintf("§[/%s]§", srcString), dependencyName, dependencyParams, nil
	}
}

func getFetch(z *html.Tokenizer, attrs []html.Attribute) (*FetchDefinition, error) {
	fd := &FetchDefinition{}

	url, hasUrl := getAttr(attrs, "src")
	if !hasUrl {
		return nil, fmt.Errorf("include definition without src %s", z.Raw())
	}
	fd.URL = strings.TrimSpace(url.Val)

	if name, hasName := getAttr(attrs, "name"); hasName {
		fd.Name = name.Val
	} else {
		fd.Name = urlToName(fd.URL)
	}

	if timeout, hasTimeout := getAttr(attrs, "timeout"); hasTimeout {
		if timeoutInt, err := strconv.Atoi(timeout.Val); err != nil {
			return nil, fmt.Errorf("error parsing timeout in %s: %s", z.Raw(), err.Error())
		} else {
			fd.Timeout = time.Millisecond * time.Duration(timeoutInt)
		}
	}

	if required, hasRequired := getAttr(attrs, "required"); hasRequired {
		if requiredBool, err := strconv.ParseBool(required.Val); err != nil {
			return nil, fmt.Errorf("error parsing bool in %s: %s", z.Raw(), err.Error())
		} else {
			fd.Required = requiredBool
		}
	}

	attr, found := getAttr(attrs, "discoveredby")
	if found {
		fd.DiscoveredBy(attr.Val)
	}

	return fd, nil
}

func ParseHeadFragment(fragment *StringFragment, headPropertyMap map[string]string) error {
	attrs := make([]html.Attribute, 0, 10)
	headBuff := bytes.NewBuffer(nil)
	z := html.NewTokenizer(strings.NewReader(string(*fragment)))
forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		raw := byteCopy(z.Raw()) // create a copy here, because readAttributes modifies z.Raw, if attributes contain an &
		attrs = readAttributes(z, attrs)

		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			break forloop
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:

			if string(tag) == "meta" {
				if processMetaTag(string(tag), attrs, headPropertyMap) {
					headBuff.Write(raw)
				}
				continue
			}
			if string(tag) == "title" {
				if headPropertyMap["title"] == "" {
					headPropertyMap["title"] = "title"
					headBuff.Write(raw)
				} else if tt != html.SelfClosingTagToken {
					skipCompleteTag(z, "title")
					continue
				}
			} else {
				headBuff.Write(raw)
			}
		default:
			headBuff.Write(raw)
		}

	}

	s := headBuff.String()

	if len(s) > 0 {
		*fragment = StringFragment(s)
	}
	return nil
}

func skipCompleteTag(z *html.Tokenizer, tagName string) error {
forloop:
	for {
		tt := z.Next()
		tag, _ := z.TagName()
		switch {
		case tt == html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			break forloop
		case tt == html.EndTagToken:
			tagAsString := string(tag)
			if tagAsString == tagName {
				break forloop
			}
		}
	}
	return nil
}

func processMetaTag(tagName string, attrs []html.Attribute, metaMap map[string]string) bool {
	if len(attrs) == 0 {
		return true
	}

	key := tagName
	value := ""
	// TODO: check explizit for attrName "http-equiv" || "name" || "charset" ?

	// e.g.: <meta charset="utf-8">
	if len(attrs) == 1 {
		key = tagName + "_" + attrs[0].Key
		value = attrs[0].Val
	}

	if len(attrs) > 1 {
		key = tagName + "_" + attrs[0].Key + "_" + attrs[0].Val
		value = attrs[1].Key + "_" + attrs[1].Val
	}

	if metaMap[key] == "" {
		metaMap[key] = value
		return true

	}
	return false
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

func skipSubtreeIfUicRemove(z *html.Tokenizer, tt html.TokenType, tagName string, attrs []html.Attribute) bool {
	_, foundRemoveTag := getAttr(attrs, UicRemove)
	if !foundRemoveTag {
		return false
	}

	if isSelfClosingTag(tagName, tt) {
		return true
	}

	depth := 0
	for {
		tt := z.Next()
		tag, _ := z.TagName()

		switch {
		case tt == html.ErrorToken:
			return true
		case tt == html.StartTagToken && !isSelfClosingTag(string(tag), tt):
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

// getFragmentName returns the name attribute, or "" if none was given
func getFragmentName(attrs []html.Attribute) string {
	for _, a := range attrs {
		if a.Key == "name" {
			return a.Val
		}
	}
	return ""
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

func joinAttrs(attrs []html.Attribute) string {
	buff := bytes.NewBufferString("")
	for i, a := range attrs {
		if i > 0 {
			buff.WriteString(" ")
		}
		if a.Namespace != "" {
			buff.WriteString(a.Namespace)
			buff.WriteString(":")
		}
		buff.WriteString(a.Key)
		buff.WriteString(`="`)
		buff.WriteString(html.EscapeString(a.Val))
		buff.WriteString(`"`)
	}
	return buff.String()
}

func isSelfClosingTag(tagname string, token html.TokenType) bool {
	return token == html.SelfClosingTagToken || voidElements[tagname]
}

// byteCopy creates a copy of a byte slice
func byteCopy(in []byte) []byte {
	result := make([]byte, len(in), len(in))
	copy(result, in)
	return result
}

// HTML Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
var voidElements = map[string]bool{
	"area":    true,
	"base":    true,
	"br":      true,
	"col":     true,
	"command": true,
	"embed":   true,
	"hr":      true,
	"img":     true,
	"input":   true,
	"keygen":  true,
	"link":    true,
	"meta":    true,
	"param":   true,
	"source":  true,
	"track":   true,
	"wbr":     true,
}
