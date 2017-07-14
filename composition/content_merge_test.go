package composition

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func Test_ContentMerge_PositiveCase(t *testing.T) {
	a := assert.New(t)

	expected := `<!DOCTYPE html>
<html>
  <head>
    <page1-head/>
    <page2-head/>
    <page3-head/>
    <link rel="stylesheet" type="text/css" href="/abc/def">
    <link rel="stylesheet" type="text/css" href="/üst/das/möglich">
  </head>
  <body a="b" foo="bar">
    <page1-body-main>
      <page2-body-a/>
      <page2-body-b/>
      <page3-body-a/>
    </page1-body-main>
    <page1-tail/>
    <page2-tail/>
  </body>
</html>
`

	body := NewStringFragment(
		`<page1-body-main>
      §[> page2-a]§
      §[> example.com#page2-b]§
      §[> page3]§
    </page1-body-main>
`)

	sheets := [][]html.Attribute{
		stylesheetAttrs("/abc/def"),
		stylesheetAttrs("/üst/das/möglich"),
	}

	body.AddLinkTags(sheets)
	cm := NewContentMerge(nil)

	cm.AddContent(&MemoryContent{
		name:           LayoutFragmentName,
		head:           NewStringFragment("<page1-head/>\n"),
		bodyAttributes: NewStringFragment(`a="b"`),
		tail:           NewStringFragment("    <page1-tail/>\n"),
		body:           map[string]Fragment{"": body},
	}, 0)

	cm.AddContent(&MemoryContent{
		name:           "example.com",
		head:           NewStringFragment("    <page2-head/>\n"),
		bodyAttributes: NewStringFragment(`foo="bar"`),
		tail:           NewStringFragment("    <page2-tail/>"),
		body: map[string]Fragment{
			"page2-a": NewStringFragment("<page2-body-a/>"),
			"page2-b": NewStringFragment("<page2-body-b/>"),
		}}, 0)

	cm.AddContent(&MemoryContent{
		name: "page3",
		head: NewStringFragment("    <page3-head/>"),
		body: map[string]Fragment{
			"": NewStringFragment("<page3-body-a/>"),
		}}, MAX_PRIORITY) // just to trigger the priority-parsing and see that it doesn't crash..

	html, err := cm.GetHtml()
	a.NoError(err)
	a.Equal(expected, string(html))
}

func Test_ContentMerge_BodyCompositionWithExplicitNames(t *testing.T) {
	a := assert.New(t)

	expected := `<!DOCTYPE html>
<html>
  <head>
    
    <link rel="stylesheet" type="text/css" href="/body/first">
    <link rel="stylesheet" type="text/css" href="/body/second">
    <link rel="stylesheet" type="text/css" href="/page/2A/first">
    <link rel="stylesheet" type="text/css" href="/page/2A/second">
    <link rel="stylesheet" type="text/css" href="/page/2B/first">
    <link rel="stylesheet" type="text/css" href="/page/2B/second">
    <link rel="stylesheet" type="text/css" href="/page/3A/first">
    <link rel="stylesheet" type="text/css" href="/page/3A/second">
  </head>
  <body>
    <page1-body-main>
      <page2-body-a/>
      <page2-body-b/>
      <page3-body-a/>
    </page1-body-main>
  </body>
</html>
`

	cm := NewContentMerge(nil)

	body := NewStringFragment(
		`<page1-body-main>
      §[> page2-a]§
      §[> example1.com#page2-b]§
      §[> page3-a]§
    </page1-body-main>`)

	sheets := [][]html.Attribute{
		stylesheetAttrs("/body/first"),
		stylesheetAttrs("/body/second"),
	}
	body.AddLinkTags(sheets)

	cm.AddContent(&MemoryContent{
		name: LayoutFragmentName,
		body: map[string]Fragment{
			"": body}}, 0)

	page2A := NewStringFragment("<page2-body-a/>")
	sheets = [][]html.Attribute{
		stylesheetAttrs("/page/2A/first"),
		stylesheetAttrs("/page/2A/second"),
	}
	page2A.AddLinkTags(sheets)

	page2B := NewStringFragment("<page2-body-b/>")
	sheets = [][]html.Attribute{
		stylesheetAttrs("/page/2B/first"),
		stylesheetAttrs("/page/2B/second"),
	}
	page2B.AddLinkTags(sheets)

	// this fragment is not rendered, so it's stylesheets should not appear in page header
	pageUnreferenced := NewStringFragment("<unreferenced-body/>")
	sheets = [][]html.Attribute{
		stylesheetAttrs("/unreferenced/first"),
		stylesheetAttrs("/unreferenced/second"),
	}
	pageUnreferenced.AddLinkTags(sheets)

	cm.AddContent(&MemoryContent{
		name: "example1.com",
		body: map[string]Fragment{
			"page2-a":      page2A,
			"page2-b":      page2B,
			"unreferenced": pageUnreferenced,
		}}, 0)

	page3A := NewStringFragment("<page3-body-a/>")
	sheets = [][]html.Attribute{
		stylesheetAttrs("/page/3A/first"),
		stylesheetAttrs("/page/3A/second"),
	}
	page3A.AddLinkTags(sheets)
	cm.AddContent(&MemoryContent{
		name: "example2.com",
		body: map[string]Fragment{
			"page3-a": page3A,
		}}, 0)

	html, err := cm.GetHtml()
	a.NoError(err)
	expected = removeTabsAndNewLines(expected)
	result := removeTabsAndNewLines(string(html))
	a.Equal(expected, result)
}

func Test_ContentMerge_LookupByDifferentFragmentNames(t *testing.T) {
	a := assert.New(t)

	fragmentA := NewStringFragment("a")
	fragmentB := NewStringFragment("b")

	cm := NewContentMerge(nil)
	cm.AddContent(&MemoryContent{
		name: "main",
		body: map[string]Fragment{
			"":  fragmentA,
			"b": fragmentB,
		}}, 0)

	// fragment a
	f, exist := cm.GetBodyFragmentByName("")
	a.True(exist)
	a.Equal(fragmentA, f)

	f, exist = cm.GetBodyFragmentByName("main")
	a.True(exist)
	a.Equal(fragmentA, f)

	f, exist = cm.GetBodyFragmentByName("main#")
	a.True(exist)
	a.Equal(fragmentA, f)

	f, exist = cm.GetBodyFragmentByName("#")
	a.True(exist)
	a.Equal(fragmentA, f)

	// fragment b
	f, exist = cm.GetBodyFragmentByName("b")
	a.True(exist)
	a.Equal(fragmentB, f)

	f, exist = cm.GetBodyFragmentByName("main#b")
	a.True(exist)
	a.Equal(fragmentB, f)

	f, exist = cm.GetBodyFragmentByName("#b")
	a.True(exist)
	a.Equal(fragmentB, f)
}

func Test_GenerateMissingFragmentString(t *testing.T) {
	body := map[string]Fragment{
		"footer": nil,
		"header": nil,
		"":       nil,
	}
	fragmentName := "body"
	fragmentString := generateMissingFragmentString(body, fragmentName)

	a := assert.New(t)
	a.Contains(fragmentString, "Fragment does not exist: body.")
	a.Contains(fragmentString, "footer")
	a.Contains(fragmentString, "header")

}

func Test_ContentMerge_MainFragmentDoesNotExist(t *testing.T) {
	a := assert.New(t)
	cm := NewContentMerge(nil)
	_, err := cm.GetHtml()
	a.Error(err)
	a.Equal("Fragment does not exist: . Existing fragments: ", err.Error())
}

type closedWriterMock struct {
}

func (buff closedWriterMock) Write(b []byte) (int, error) {
	return 0, errors.New("writer closed")
}

func asFetchResult(c Content) *FetchResult {
	return &FetchResult{Content: c, Def: &FetchDefinition{URL: c.Name()}}
}

func Test_MergeMultipleContents(t *testing.T) {
	a := assert.New(t)

	cm := NewContentMerge(nil)
	cm.AddContent(getFixtureWithName(LayoutFragmentName, "simple/layout1.html"), 0)
	cm.AddContent(getFixture("simple/fragment_header.html"), 0)
	cm.AddContent(getFixture("simple/fragment_content.html"), 0)
	cm.AddContent(getFixture("simple/fragment_header2.html"), 0)

	html, err := cm.GetHtml()
	a.NoError(err)
	sHtml := string(html)
	a.Contains(sHtml, "TEST-CONTENT")
	a.Contains(sHtml, "TEST-HEADER 2")
	a.Contains(sHtml, "<title>layout-header</title>")
	a.Contains(sHtml, "<title>test-header</title>")
	a.Contains(sHtml, "<title>content-header</title>")
	a.Contains(sHtml, "<title>test-header 2</title>")
}

func Test_MergeMultipleContentsPrioritized(t *testing.T) {
	a := assert.New(t)

	cm := NewContentMerge(nil)
	cm.AddContent(getFixtureWithName(LayoutFragmentName, "simple/layout1.html"), 0)
	cm.AddContent(getFixture("simple/fragment_header.html"), 0)
	cm.AddContent(getFixture("simple/fragment_content.html"), 1)
	cm.AddContent(getFixture("simple/fragment_header2.html"), 0)

	html, err := cm.GetHtml()
	a.NoError(err)
	sHtml := string(html)
	a.Contains(sHtml, "TEST-CONTENT")
	a.Contains(sHtml, "TEST-HEADER 2")
	// Notice: This assertion is somewhat unexpected. Normally one would expect
	// the title of fragment_content.html here, which is given a higher priority.
	// But prioritization is done somewhere else in this library and the
	// priority value of the AddContent() method is only used as a flag.
	a.Contains(sHtml, "<title>test-header 2</title>")
}

func Test_MergeScriptTags(t *testing.T) {
	a := assert.New(t)

	cm := NewContentMerge(nil)
	cm.AddContent(getFixtureWithName(LayoutFragmentName, "with_scripts/layout_scripts.html"), 0)
	cm.AddContent(getFixture("with_scripts/fragment_scripts_header.html"), 0)
	cm.AddContent(getFixture("with_scripts/fragment_scripts_footer.html"), 1)
	cm.AddContent(getFixture("with_scripts/fragment_scripts_content.html"), 0)

	expected, err := ioutil.ReadFile("testdata/with_scripts/expected_scripts.html")
	sExpected := trim(string(expected))
	a.NoError(err)

	html, err := cm.GetHtml()
	a.NoError(err)
	sHtml := trim(string(html))
	a.Equal(sHtml, sExpected)
}

func Test_MergeScriptTagsWithPrios(t *testing.T) {
	a := assert.New(t)

	cm := NewContentMerge(nil)
	cm.AddContent(getFixtureWithName(LayoutFragmentName, "with_scripts/layout_scripts.html"), 0)
	cm.AddContent(getFixture("with_scripts/fragment_scripts_header.html"), 0)
	cm.AddContent(getFixture("with_scripts/fragment_scripts_footer.html"), 1)
	cm.AddContent(getFixture("with_scripts/fragment_scripts_content.html"), 0)

	expected, err := ioutil.ReadFile("testdata/with_scripts/expected_scripts.html")
	sExpected := trim(string(expected))
	a.NoError(err)

	html, err := cm.GetHtml()
	a.NoError(err)
	sHtml := trim(string(html))
	a.Equal(sHtml, sExpected)
}

func trim(html string) string {
	splitted := strings.Split(html, "\n")
	var result []string
	for _, v := range splitted {
		trimmed := strings.Trim(v, " \t\n\r")
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}

func getFixtureWithName(name string, filename string) (c *MemoryContent) {
	c = getFixture(filename)
	c.name = name
	return c
}

func getFixture(filename string) (c *MemoryContent) {
	dat, err := ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		panic(err)
	}
	c, err = parse(string(dat))
	if err != nil {
		panic(err)
	}
	return c
}

func parse(buf string) (c *MemoryContent, err error) {
	parser := NewContentParser(true, true)
	z := bytes.NewBufferString(buf)
	c = NewMemoryContent()
	err = parser.Parse(c, z)
	return c, err
}
