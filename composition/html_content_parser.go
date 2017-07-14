package composition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type HtmlContentParser struct {
	collectLinks   bool
	collectScripts bool
}

func NewHtmlContentParser(collectLinks bool, collectScripts bool) *HtmlContentParser {
	return &HtmlContentParser{collectLinks: collectLinks, collectScripts: collectScripts}
}

const (
	UicRemove       = "uic-remove"
	UicInclude      = "uic-include"
	UicFetch        = "uic-fetch"
	UicFragment     = "uic-fragment"
	UicTail         = "uic-tail"
	ScriptTypeMeta  = "text/uic-meta"
	ParamAttrPrefix = "param-"
)

type TagType int

const (
	LINK TagType = iota
	META
	SCRIPT
	SCRIPT_INLINE
	UNKNOWN
)

func getTag(tag []byte, attrs []html.Attribute) (tagAttrs []html.Attribute, tagType TagType) {
	tagAttrs = nil
	tagAttrs = append(tagAttrs, attrs...)
	if string(tag) == "link" {
		return tagAttrs, LINK
	}
	if string(tag) == "meta" {
		return tagAttrs, META
	}
	if string(tag) == "script" {
		if _, hasUrl := getAttr(attrs, "src"); !hasUrl {
			return tagAttrs, SCRIPT_INLINE
		}
		return tagAttrs, SCRIPT
	}
	return nil, UNKNOWN
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

func (parser *HtmlContentParser) collectLinksAndScripts(tag []byte, attrs []html.Attribute, linkTags *[][]html.Attribute, scriptTags *[]ScriptElement, z *html.Tokenizer, tt html.TokenType) (skip bool, err error) {

	skip = false
	tagAttrs, tagType := getTag(tag, attrs)
	if tagType == UNKNOWN {
		// do nothing
	} else if tagType == LINK && parser.collectLinks {
		*linkTags = append(*linkTags, tagAttrs)
		skip = true
	} else if tagType == SCRIPT && parser.collectScripts {
		*scriptTags = append(*scriptTags, newScriptElement(tagAttrs, nil))
		if skipSubtree(z, tt, string(tag), attrs) {
			skip = true
		}
	} else if tagType == SCRIPT_INLINE && parser.collectScripts {
		txt, err := parseInlineScript(z)
		if err != nil {
			return false, err
		}
		*scriptTags = append(*scriptTags, newScriptElement(tagAttrs, txt))
		skip = true
	}
	return skip, nil
}

func nextToken(z *html.Tokenizer, attrs []html.Attribute, stopToken ...string) (error, html.TokenType, []byte, []byte, []html.Attribute) {
	tt := z.Next()
	tag, _ := z.TagName()
	raw := byteCopy(z.Raw()) // create a copy here, because readAttributes modifies z.Raw, if attributes contain an &
	attrs = readAttributes(z, attrs)

	switch {

	case tt == html.ErrorToken:
		return z.Err(), tt, tag, raw, attrs

	case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
		if skipSubtreeIfUicRemove(z, tt, string(tag), attrs) {
			return nextToken(z, attrs, stopToken...)
		}

	case tt == html.EndTagToken:
		for _, tok := range stopToken {
			if string(tag) == tok {
				return io.EOF, tt, tag, raw, attrs
			}
		}
	}

	// return the next parseable token
	return nil, tt, tag, raw, attrs
}

func (parser *HtmlContentParser) parseHead(z *html.Tokenizer, c *MemoryContent) error {
	var linkTags [][]html.Attribute
	var scriptTags []ScriptElement
	attrs := make([]html.Attribute, 0, 10)
	headBuff := bytes.NewBuffer(nil)

forloop:
	for {
		err, tt, tag, raw, attrs := nextToken(z, attrs, "head")
		if err != nil {
			if err != io.EOF {
				return z.Err()
			}
			break forloop
		}

		if string(tag) == "script" && attrHasValue(attrs, "type", ScriptTypeMeta) {
			if err := parseMetaJson(z, c); err != nil {
				return err
			}
			continue
		}

		skip, err := parser.collectLinksAndScripts(tag, attrs, &linkTags, &scriptTags, z, tt)
		if err != nil {
			return err
		}

		if skip {
			continue
		}
		headBuff.Write(raw)
	}

	s := headBuff.String()
	st := strings.Trim(s, " \n")
	if len(st) > 0 || len(linkTags) > 0 || len(scriptTags) > 0 {
		frg := NewStringFragment(st)
		frg.AddLinkTags(linkTags)
		frg.AddScriptTags(scriptTags)
		c.head = frg
	}
	return nil
}

func (parser *HtmlContentParser) parseBody(z *html.Tokenizer, c *MemoryContent) error {
	var linkTags [][]html.Attribute
	var scriptTags []ScriptElement

	attrs := make([]html.Attribute, 0, 10)
	bodyBuff := bytes.NewBuffer(nil)

	attrs = readAttributes(z, attrs)
	if len(attrs) > 0 {
		c.bodyAttributes = NewStringFragment(joinAttrs(attrs))
	}

forloop:
	for {
		err, tt, tag, raw, attrs := nextToken(z, attrs, "body")
		if err != nil {
			if err != io.EOF {
				return z.Err()
			}
			break forloop
		}

		switch {
		case tt == html.StartTagToken || tt == html.SelfClosingTagToken:
			if string(tag) == UicFragment {
				if f, deps, err := parser.parseFragment(z); err != nil {
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
				if f, deps, err := parser.parseFragment(z); err != nil {
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

			skip, err := parser.collectLinksAndScripts(tag, attrs, &linkTags, &scriptTags, z, tt)
			if err != nil {
				return err
			}

			if skip {
				continue
			}
		}
		bodyBuff.Write(raw)
	}

	s := bodyBuff.String()
	if _, defaultFragmentExists := c.body[""]; !defaultFragmentExists {
		if st := strings.Trim(s, " \n"); len(st) > 0 || len(linkTags) > 0 || len(scriptTags) > 0 {
			frg := NewStringFragment(st)
			frg.AddLinkTags(linkTags)
			frg.AddScriptTags(scriptTags)
			c.body[""] = frg
		}
	}

	return nil
}

func (parser *HtmlContentParser) parseFragment(z *html.Tokenizer) (f Fragment, dependencies map[string]Params, err error) {
	var linkTags [][]html.Attribute
	var scriptTags []ScriptElement
	attrs := make([]html.Attribute, 0, 10)
	dependencies = make(map[string]Params)

	buff := bytes.NewBuffer(nil)
forloop:
	for {
		err, tt, tag, raw, attrs := nextToken(z, attrs, UicFragment, UicTail)
		if err != nil {
			if err != io.EOF {
				return nil, nil, z.Err()
			}
			break forloop
		}

		switch {
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

			skip, err := parser.collectLinksAndScripts(tag, attrs, &linkTags, &scriptTags, z, tt)
			if err != nil {
				return nil, nil, err
			}

			if skip {
				continue
			}

		}
		buff.Write(raw)
	}

	frg := NewStringFragment(buff.String())
	frg.AddLinkTags(linkTags)
	frg.AddScriptTags(scriptTags)
	return frg, dependencies, nil
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
	z := html.NewTokenizer(strings.NewReader(fragment.Content()))
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

			switch {
			case string(tag) == "meta":
				if processMetaTag(string(tag), attrs, headPropertyMap) {
					headBuff.Write(raw)
				}
				continue forloop
			case string(tag) == "link":
				if processLinkTag(attrs, headPropertyMap) {
					headBuff.Write(raw)
				}
				continue forloop
			case string(tag) == "title":
				if headPropertyMap["title"] == "" {
					headPropertyMap["title"] = "title"
					headBuff.Write(raw)
				} else if tt != html.SelfClosingTagToken {
					skipCompleteTag(z, "title")
				}
				continue forloop
			default:
				headBuff.Write(raw)
			}

		default:
			headBuff.Write(raw)
		}

	}

	s := headBuff.String()

	if len(s) > 0 {
		fragment.SetContent(s)
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

	// e.g.: <meta charset="utf-8"> => key = meta_charset; val = utf-8
	if len(attrs) == 1 {
		key = tagName + "_" + attrs[0].Key
		value = attrs[0].Val
	}

	// e.g.: <meta name="content-language" content="de"> => key = meta_name_content-language; val = content_de
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

/**
Returns true if a link tag can be processed.
Checks if a <link> tag contains a canonical relation and avoids multiple canonical definitions.
*/
func processLinkTag(attrs []html.Attribute, metaMap map[string]string) bool {
	if len(attrs) == 0 {
		return true
	}

	const canonical = "canonical"
	var key string
	var value string

	// e.g.: <link rel="canonical" href="/baumarkt/suche"> => key = canonical; val = /baumarkt/suche
	for _, attr := range attrs {
		if attr.Key == "rel" && attr.Val == canonical {
			key = canonical
		}
		if attr.Key == "href" {
			value = attr.Val
		}
	}
	if key == canonical && metaMap[canonical] != "" {
		// if canonical is already in map then don't process this link tag
		return false
	}

	if key != "" && value != "" {
		metaMap[key] = value
	}
	return true
}

func parseInlineScript(z *html.Tokenizer) ([]byte, error) {
	tt := z.Next()
	if tt != html.TextToken {
		tag, _ := z.TagName()
		if tt == html.EndTagToken && string(tag) == "script" {
			return nil, nil // don't treat empty scripts as error
		}
		return nil, fmt.Errorf("expected text node for inline script, but found %v, (%s)", tt.String(), z.Raw())
	}

	bytes := z.Text()

	tt = z.Next()
	tag, _ := z.TagName()
	if tt != html.EndTagToken || string(tag) != "script" {
		msg := "Tag not properly ended. Expected </script>"
		if tag != nil {
			msg = msg + ", but found " + string(tag)
		}
		if tt == html.ErrorToken {
			msg = msg + ". Error was: " + z.Err().Error()
		}
		return nil, fmt.Errorf(msg)
	}

	return bytes, nil
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
	// return skipSubtree(z, tt, tagName, attrs)
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

func skipSubtree(z *html.Tokenizer, tt html.TokenType, tagName string, attrs []html.Attribute) bool {
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
