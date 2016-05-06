package composition

import (
	"bytes"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func Test_ContentMerge_PositiveCase(t *testing.T) {
	a := assert.New(t)

	expected := `<html>
  <head>
    <page1-head>
    <page2-head>
    <page3-head>
  </head>
  <body>
    <page1-body-main>
      <page2-body-a>
      <page2-body-b>
      <page3-body-a>
    </page1-body-main>
    <page1-tail>
    <page2-tail>
  </body>
</html>
`

	page1 := NewMemoryContent()
	page1.head = StringFragment("    <page1-head>\n")
	page1.tail = StringFragment("    <page1-tail>\n")
	page1.body[""] = MockPage1BodyFragment{}

	page2 := NewMemoryContent()
	page2.head = StringFragment("    <page2-head>\n")
	page2.tail = StringFragment("    <page2-tail>\n")
	page2.body["page2-a"] = StringFragment("      <page2-body-a>\n")
	page2.body["page2-b"] = StringFragment("      <page2-body-b>\n")

	page3 := NewMemoryContent()
	page3.head = StringFragment("    <page3-head>\n")
	page3.body["page3-a"] = StringFragment("      <page3-body-a>\n")

	cm := NewContentMerge()
	cm.AddContent(page1)
	cm.AddContent(page2)
	cm.AddContent(page3)

	buff := bytes.NewBuffer(nil)
	err := cm.WriteHtml(buff)

	a.NoError(err)
	a.Equal(expected, buff.String())
}

func Test_ContentMerge_MetadataIsMerged_And_SuppliedToFragments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	expectedMeta := map[string]interface{}{
		"page1": "value1",
		"page2": "value2",
	}
	bodyMock := NewMockFragment(ctrl)
	bodyMock.EXPECT().Execute(gomock.Any(), expectedMeta, gomock.Any())

	page1 := NewMemoryContent()
	page1.meta["page1"] = "value1"
	page1.body[""] = bodyMock

	page2 := NewMemoryContent()
	page2.meta["page2"] = "value2"

	cm := NewContentMerge()
	cm.AddContent(page1)
	cm.AddContent(page2)

	err := cm.WriteHtml(bytes.NewBuffer(nil))
	a.NoError(err)
}

func Test_ContentMerge_MainFragmentDoesNotExist(t *testing.T) {
	a := assert.New(t)
	cm := NewContentMerge()
	buff := bytes.NewBuffer(nil)
	err := cm.WriteHtml(buff)
	a.Error(err)
	a.Equal("Fragment does not exist: ", err.Error())
	// the buffered merger should not write if errors occur
	a.Equal(0, len(buff.Bytes()))
}

func Test_ContentMerge_ErrorOnWrite(t *testing.T) {
	a := assert.New(t)

	page := NewMemoryContent()
	page.body[""] = StringFragment("Hello World\n")

	cm := NewContentMerge()
	cm.AddContent(page)

	err := cm.WriteHtml(closedWriterMock{})
	a.Error(err)
	a.Equal("writer closed", err.Error())
}

func Test_ContentMerge_ErrorOnWriteUnbuffered(t *testing.T) {
	a := assert.New(t)

	cm := NewContentMerge()
	cm.Buffered = false
	err := cm.WriteHtml(closedWriterMock{})
	a.Error(err)
	a.Equal("writer closed", err.Error())
}

type MockPage1BodyFragment struct {
}

func (f MockPage1BodyFragment) Execute(w io.Writer, data map[string]interface{}, executeNestedFragment func(nestedFragmentName string) error) error {
	w.Write([]byte("    <page1-body-main>\n"))
	if err := executeNestedFragment("page2-a"); err != nil {
		panic(err)
	}
	if err := executeNestedFragment("page2-b"); err != nil {
		panic(err)
	}
	if err := executeNestedFragment("page3-a"); err != nil {
		panic(err)
	}
	w.Write([]byte("    </page1-body-main>\n"))
	return nil
}

type closedWriterMock struct {
}

func (buff closedWriterMock) Write(b []byte) (int, error) {
	return 0, errors.New("writer closed")
}
