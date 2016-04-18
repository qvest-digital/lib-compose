package aggregation

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testhtml = `<html>
  <head>
    <link uia-remove rel="stylesheet" type="text/css" href="testing.css">
    <link rel="stylesheet" type="text/css" href="special.css">
    <script type="text/uia-meta">
      {
       "foo": "bar",
       "boo": "bazz",
       "categories": ["animal", "human"]
      }
    </script>
    <script uia-remove>
      va xzw = "some test code""
    </script>
    <script src="myScript.js"></script>
  </head>
  <body>
    <ul uia-remove>
      <!-- A Navigation for testing -->
    </ul>
    <div uia-fragment="headline">
      <h1>This is a headline</h1>
    </div>
    <div uia-fragment="content">
      Bli Bla blub
      <uia-include src="example.com/foo#content" timeout="42000" required="true"/>
      <uia-include src="example.com/optional#content" timeout="100" required="false"/>
      <div uia-remove>
         Some element for testing
      </div>
      <hr/>
      Bli Bla blub
    </div>
    <div uia-tail>
      <!-- some script tags to insert at the end -->
      <script src="foo.js"></script>
      <script src="bar.js"></script>
      <script uia-remove src="demo.js"></script>
    </div>
  </body>
</html>
`

func Test_HtmlContentLoader_LoadError500(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", 500)
	}))
	defer server.Close()

	loader := &HtmlContentLoader{}
	c, err := loader.Load(server.URL, time.Second)
	a.Error(err)
	a.Nil(c)
	a.Contains(err.Error(), "http 500")
}

func Test_HtmlContentLoader_LoadErrorNetwork(t *testing.T) {
	a := assert.New(t)

	loader := &HtmlContentLoader{}
	c, err := loader.Load("...", time.Second)
	a.Error(err)
	a.Nil(c)
	a.Contains(err.Error(), "unsupported protocol scheme")
}

func Test_HtmlContentLoader_LoadEmptyContent(t *testing.T) {
	a := assert.New(t)

	server := testServer(`<html>
  <head>
  </head>
  <body>
  </body>
</html>
`, time.Millisecond*0)
	defer server.Close()

	loader := &HtmlContentLoader{}
	c, err := loader.Load(server.URL, time.Second)
	a.NoError(err)
	a.NotNil(c)

	a.Equal(0, len(c.Body()))
	a.Equal(0, len(c.Meta()))
	a.Equal(0, len(c.RequiredContent()))
	a.Nil(c.Head())
	a.Nil(c.Tail())
}

func Test_HtmlContentLoader_parseHead(t *testing.T) {
	a := assert.New(t)

	loader := &HtmlContentLoader{}
	z := html.NewTokenizer(bytes.NewBufferString(`
<div uia-remove>
  <script>
  sdcsdc
  </script>
</div>
<xx/> 
<foo>xxx</foo>
<div uia-remove>
  <script>
  sdcsdc
  </script>
</div> 
<bar>xxx</bar>
<script type="text/uia-meta">
      {
       "foo": "bar"
      }
</script>
<div uia-remove>
  <script>
  sdcsdc
  </script>
</div> 
`))

	c := NewMemoryContent()
	err := loader.parseHead(z, c)
	a.NoError(err)

	eqFragment(t, "<xx/><foo>xxx</foo><bar>xxx</bar>", c.Head())
	a.Equal("bar", c.Meta()["foo"])
}

func Test_HtmlContentLoader_parseBody(t *testing.T) {
	a := assert.New(t)

	loader := &HtmlContentLoader{}
	z := html.NewTokenizer(bytes.NewBufferString(`<body>
    <ul uia-remove>
      <!-- A Navigation for testing -->
    </ul>
    <div uia-fragment="headline">
      <h1>Headline</h1>
      <uia-include src="example.com/optional#content" timeout="100" required="false"/>
    </div>
    <div uia-fragment="content">
      some content
      <uia-include src="example.com/foo#content" timeout="42000" required="true"/>
      <uia-include src="example.com/optional#content" timeout="100" required="false"/>
    </div>
    <div uia-tail>
      <!-- tail -->
    </div>
  </body>`))

	z.Next() // At <body ..
	c := NewMemoryContent()
	err := loader.parseBody(z, c)
	a.NoError(err)

	a.Equal(2, len(c.Body()))
	headline := c.Body()["headline"]
	eqFragment(t, `<div uia-fragment="headline">
      <h1>Headline</h1>
      §[> example.com/optional#content]§
    </div>`, headline)

	content := c.Body()["content"]
	eqFragment(t, `<div uia-fragment="content">some content §[> example.com/foo#content]§ §[> example.com/optional#content]§</div>`, content)

	eqFragment(t, "<div uia-tail><!-- tail --></div>", c.Tail())

	a.Equal(2, len(c.RequiredContent()))
	a.Equal(&FetchDefinition{
		URL:      "example.com/foo#content",
		Timeout:  time.Millisecond * 42000,
		Required: true,
	}, c.requiredContent["example.com/foo#content"])

	a.Equal(&FetchDefinition{
		URL:      "example.com/optional#content",
		Timeout:  time.Millisecond * 100,
		Required: false,
	}, c.requiredContent["example.com/optional#content"])
}

func Test_HtmlContentLoader_parseHead_JsonError(t *testing.T) {
	a := assert.New(t)

	loader := &HtmlContentLoader{}
	z := html.NewTokenizer(bytes.NewBufferString(`
<script type="text/uia-meta">
      {
</script>
`))

	c := NewMemoryContent()
	err := loader.parseHead(z, c)

	a.Error(err)
	a.Contains(err.Error(), "error while parsing json from meta json")
}

func Test_HtmlContentLoader_parseFragment(t *testing.T) {
	a := assert.New(t)

	z := html.NewTokenizer(bytes.NewBufferString(`<div uia-fragment="content">
      Bli Bla blub
      <uia-include src="example.com/foo#content" timeout="42000" required="true"/>
      <uia-include src="example.com/optional#content" timeout="100" required="false"/>
      <div uia-remove>
         Some element for testing
      </div>
      <hr/>     
      Bli Bla §[ aVariable ]§ blub
    </div>`))

	z.Next() // At <div uia-fragment ..
	f, deps, err := parseFragment(z)
	a.NoError(err)

	a.Equal(2, len(deps))
	a.Equal(&FetchDefinition{
		URL:      "example.com/foo#content",
		Timeout:  time.Millisecond * 42000,
		Required: true,
	}, deps[0])

	a.Equal(&FetchDefinition{
		URL:      "example.com/optional#content",
		Timeout:  time.Millisecond * 100,
		Required: false,
	}, deps[1])

	sFragment := f.(StringFragment)
	expected := `<div uia-fragment="content">
      Bli Bla blub
      §[> example.com/foo#content]§
      §[> example.com/optional#content]§
      <hr/>
      Bli Bla §[ aVariable ]§ blub
    </div>`
	eqFragment(t, expected, sFragment)
}

func Test_HtmlContentLoader_parseMetaJson(t *testing.T) {
	a := assert.New(t)

	z := html.NewTokenizer(bytes.NewBufferString(`<script type="text/uia-meta">
      {
       "foo": "bar",
       "boo": "bazz",
       "categories": ["animal", "human"]
      }
    </script>`))

	z.Next() // At <script ..
	c := NewMemoryContent()
	err := parseMetaJson(z, c)
	a.NoError(err)

	a.Equal("bar", c.Meta()["foo"])
}

func Test_HtmlContentLoader_parseMetaJson_Errors(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		html      string
		errorText string
	}{
		{
			html:      `<script type="text/uia-meta"></script>`,
			errorText: "expected text node for meta",
		},
		{
			html:      `<script type="text/uia-meta">{"sdc":</script>`,
			errorText: "error while parsing json from meta json",
		},
		{
			html:      `<script type="text/uia-meta">{}`,
			errorText: "Tag not properly ended",
		},
	}

	for _, test := range testCases {
		z := html.NewTokenizer(bytes.NewBufferString(test.html))
		z.Next() // At <script ..
		err := parseMetaJson(z, NewMemoryContent())

		a.Error(err)
		a.Contains(err.Error(), test.errorText)
	}
}

func Test_HtmlContentLoader_skipSubtreeIfUiaRemove(t *testing.T) {
	a := assert.New(t)

	z := html.NewTokenizer(bytes.NewBufferString(`<a><b uia-remove>
    sdcsdc
    <foo>xxx</foo>
    <br/>
</b></a>`))

	z.Next()
	tt := z.Next() // at b
	attrs := readAttributes(z, make([]html.Attribute, 0, 10))
	skipped := skipSubtreeIfUiaRemove(z, tt, "b", attrs)

	a.True(skipped)
	token := z.Next()
	a.Equal(html.EndTagToken, token)
	tag, _ := z.TagName()
	a.Equal("a", string(tag))
}

func testServer(content string, timeout time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout)
		w.Write([]byte(content))
	}))
}

func eqFragment(t *testing.T, expected string, f Fragment) {
	if f == nil {
		t.Error("Fragment is nil, but expected:", expected)
	}
	sf := f.(StringFragment)
	sfStripped := strings.Replace(string(sf), " ", "", -1)
	sfStripped = strings.Replace(string(sfStripped), "\n", "", -1)
	expectedStripped := strings.Replace(expected, " ", "", -1)
	expectedStripped = strings.Replace(expectedStripped, "\n", "", -1)

	if expectedStripped != sfStripped {
		t.Error("Fragment is not equal: \nexpected: ", expected, "\nactual:  ", sf)
	}
}
