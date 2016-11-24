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
      §[> page3-a]§
    </page1-body-main>
`)

	cm := NewContentMerge(nil)

	cm.AddContent(&MemoryContent{
		name:           "example.com",
		head:           StringFragment("<page1-head/>\n"),
		bodyAttributes: StringFragment(`a="b"`),
		tail:           StringFragment("    <page1-tail/>\n"),
		body:           map[string]Fragment{"": body},
	}, 0)

	page2 := NewMemoryContent()
	page2.name = "example.com"
	page2.head = StringFragment("    <page2-head/>\n")
	page2.bodyAttributes = StringFragment(`foo="bar"`)
	page2.tail = StringFragment("    <page2-tail/>")
	page2.body["page2-a"] = StringFragment("<page2-body-a/>")
	page2.body["page2-b"] = StringFragment("<page2-body-b/>")

	page3 := NewMemoryContent()
	page3.name = "example.com"
	page3.head = StringFragment("    <page3-head/>")
	page3.body["page3-a"] = StringFragment("<page3-body-a/>")

	cm.AddContent(page2, 0)
	cm.AddContent(page3, 0)

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

	page1 := NewMemoryContent()
	page1.name = LayoutFragmentName
	page1.body[""] = StringFragment(
		`<page1-body-main>
      §[> page2-a]§
      §[> example.com#page2-b]§
      §[> page3-a]§
    </page1-body-main>`)

	page2 := NewMemoryContent()
	page2.name = "example.com"
	page2.body["page2-a"] = StringFragment("<page2-body-a/>")
	page2.body["page2-b"] = StringFragment("<page2-body-b/>")

	page3 := NewMemoryContent()
	page3.name = "example.com"
	page3.body["page3-a"] = StringFragment("<page3-body-a/>")

	cm := NewContentMerge(nil)
	cm.AddContent(page1, 0)
	cm.AddContent(page2, 0)
	cm.AddContent(page3, 0)

	html, err := cm.GetHtml()
	a.NoError(err)
	a.Equal(expected, string(html))
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

func Test_ContentMerge_FdHashes(t *testing.T) {
	a := assert.New(t)
	cm := NewContentMerge(nil)

	cm.addFdHash("testHash")
	a.Equal(cm.GetHashes()[0], "testHash")
}

type closedWriterMock struct {
}

func (buff closedWriterMock) Write(b []byte) (int, error) {
	return 0, errors.New("writer closed")
}

func asFetchResult(c Content) *FetchResult {
	return &FetchResult{Content: c, Def: &FetchDefinition{URL: c.Name()}}
}
