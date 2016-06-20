package composition

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_ContentMerge_PositiveCase(t *testing.T) {
	a := assert.New(t)

	expected := `<html>
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

	page1 := NewMemoryContent()
	page1.url = "example.com"
	page1.head = StringFragment("<page1-head/>\n")
	page1.bodyAttributes = StringFragment(`a="b"`)
	page1.tail = StringFragment("    <page1-tail/>\n")
	page1.body[""] = MockPage1BodyFragment{}

	page2 := NewMemoryContent()
	page2.url = "example.com"
	page2.head = StringFragment("    <page2-head/>\n")
	page2.bodyAttributes = StringFragment(`foo="bar"`)
	page2.tail = StringFragment("    <page2-tail/>")
	page2.body["page2-a"] = StringFragment("      <page2-body-a/>\n")
	page2.body["page2-b"] = StringFragment("      <page2-body-b/>\n")

	page3 := NewMemoryContent()
	page3.url = "example.com"
	page3.head = StringFragment("    <page3-head/>")
	page3.body["page3-a"] = StringFragment("      <page3-body-a/>\n")

	cm := NewContentMerge(nil)
	cm.AddContent(asFetchResult(page1))
	cm.AddContent(asFetchResult(page2))
	cm.AddContent(asFetchResult(page3))

	html, err := cm.GetHtml()
	a.NoError(err)
	a.Equal(expected, string(html))
}

type MockPage1BodyFragment struct {
}

func (f MockPage1BodyFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	w.Write([]byte("<page1-body-main>\n"))
	if err := executeNestedFragment("page2-a"); err != nil {
		panic(err)
	}
	if err := executeNestedFragment("example.com#page2-b"); err != nil {
		panic(err)
	}
	if err := executeNestedFragment("page3-a"); err != nil {
		panic(err)
	}
	w.Write([]byte("    </page1-body-main>\n"))
	return nil
}

func Test_ContentMerge_MainFragmentDoesNotExist(t *testing.T) {
	a := assert.New(t)
	cm := NewContentMerge(nil)
	_, err := cm.GetHtml()
	a.Error(err)
	a.Equal("Fragment does not exist: ", err.Error())
}

type closedWriterMock struct {
}

func (buff closedWriterMock) Write(b []byte) (int, error) {
	return 0, errors.New("writer closed")
}

func asFetchResult(c Content) *FetchResult {
	return &FetchResult{Content: c, Def: &FetchDefinition{URL: c.URL()}}
}
