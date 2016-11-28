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

	body := StringFragment(
		`<page1-body-main>
      §[> page2-a]§
      §[> example.com#page2-b]§
      §[> page3]§
    </page1-body-main>
`)

	cm := NewContentMerge(nil)

	cm.AddContent(&MemoryContent{
		name:           LayoutFragmentName,
		head:           StringFragment("<page1-head/>\n"),
		bodyAttributes: StringFragment(`a="b"`),
		tail:           StringFragment("    <page1-tail/>\n"),
		body:           map[string]Fragment{"": body},
	}, 0)

	cm.AddContent(&MemoryContent{
		name:           "example.com",
		head:           StringFragment("    <page2-head/>\n"),
		bodyAttributes: StringFragment(`foo="bar"`),
		tail:           StringFragment("    <page2-tail/>"),
		body: map[string]Fragment{
			"page2-a": StringFragment("<page2-body-a/>"),
			"page2-b": StringFragment("<page2-body-b/>"),
		}}, 0)

	cm.AddContent(&MemoryContent{
		name: "page3",
		head: StringFragment("    <page3-head/>"),
		body: map[string]Fragment{
			"": StringFragment("<page3-body-a/>"),
		}}, 0)

	html, err := cm.GetHtml()
	a.NoError(err)
	a.Equal(expected, string(html))
}

func Test_ContentMerge_BodyCompositionWithExplicitNames(t *testing.T) {
	a := assert.New(t)

	expected := `<!DOCTYPE html>
<html>
  <head>
    
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

	cm.AddContent(&MemoryContent{
		name: LayoutFragmentName,
		body: map[string]Fragment{
			"": StringFragment(
				`<page1-body-main>
      §[> page2-a]§
      §[> example1.com#page2-b]§
      §[> page3-a]§
    </page1-body-main>`)}}, 0)

	cm.AddContent(&MemoryContent{
		name: "example1.com",
		body: map[string]Fragment{
			"page2-a": StringFragment("<page2-body-a/>"),
			"page2-b": StringFragment("<page2-body-b/>"),
		}}, 0)

	cm.AddContent(&MemoryContent{
		name: "example2.com",
		body: map[string]Fragment{
			"page3-a": StringFragment("<page3-body-a/>"),
		}}, 0)

	html, err := cm.GetHtml()
	a.NoError(err)
	a.Equal(expected, string(html))
}

func Test_ContentMerge_LookupByDifferentFragmentNames(t *testing.T) {
	a := assert.New(t)

	fragmentA := StringFragment("a")
	fragmentB := StringFragment("b")

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
