package composition

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
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

	body.AddStylesheets([]string{"/abc/def", "/üst/das/möglich"})
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
	body.AddStylesheets([]string{"/body/first", "/body/second"})
	cm.AddContent(&MemoryContent{
		name: LayoutFragmentName,
		body: map[string]Fragment{
			"": body}}, 0)

	page2A := NewStringFragment("<page2-body-a/>")
	page2A.AddStylesheets([]string{"/page/2A/first", "/page/2A/second"})
	page2B := NewStringFragment("<page2-body-b/>")
	page2B.AddStylesheets([]string{"/page/2B/first", "/page/2B/second"})

	// this fragment is not rendered, so it's stylesheets should not appear in page header
	pageUnreferenced := NewStringFragment("<unreferenced-body/>")
	pageUnreferenced.AddStylesheets([]string{"/unreferenced/first", "/unreferenced/second"})

	cm.AddContent(&MemoryContent{
		name: "example1.com",
		body: map[string]Fragment{
			"page2-a": page2A,
			"page2-b": page2B,
			"unreferenced": pageUnreferenced,
		}}, 0)

	page3A := NewStringFragment("<page3-body-a/>")
	page3A.AddStylesheets([]string{"/page/3A/first", "/page/3A/second"})
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
