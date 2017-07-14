package composition

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func Test_StringFragment(t *testing.T) {
	a := assert.New(t)

	f := NewStringFragment("§[foo]§")
	sheets := [][]html.Attribute{
		[]html.Attribute{{Key: "href", Val: "/abc/def"}},
		[]html.Attribute{{Key: "href", Val: "ghi/xyz"}}}

	f.AddLinkTags(sheets)
	a.EqualValues(sheets, f.LinkTags())
	buf := bytes.NewBufferString("")
	err := f.Execute(buf, map[string]interface{}{"foo": "bar"}, nil)
	a.NoError(err)

	a.Equal("bar", buf.String())
}
