package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func XTest_HttpContentLoader_Load(t *testing.T) {
	a := assert.New(t)

	server := testServer("", time.Millisecond*0)
	defer server.Close()

	loader := &HttpContentLoader{}
	c, err := loader.Load(NewFetchDefinition(server.URL))
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())

	a.Equal(server.URL, c.URL())
	eqFragment(t, integratedTestHtmlExpectedHead, c.Head())
	a.Equal(2, len(c.Body()))

	eqFragment(t, integratedTestHtmlExpectedHeadline, c.Body()["headline"])
	eqFragment(t, integratedTestHtmlExpectedContent, c.Body()["content"])
	a.Equal(integratedTestHtmlExpectedMeta, c.Meta())
	eqFragment(t, integratedTestHtmlExpectedTail, c.Tail())
	cMemoryConent := c.(*MemoryContent)
	a.Equal(2, len(cMemoryConent.RequiredContent()))
	a.Equal(&FetchDefinition{
		URL:      "example.com/foo",
		Timeout:  time.Millisecond * 42000,
		Required: true,
	}, cMemoryConent.requiredContent["example.com/foo"])

	a.Equal(&FetchDefinition{
		URL:      "example.com/optional",
		Timeout:  time.Millisecond * 100,
		Required: false,
	}, cMemoryConent.requiredContent["example.com/optional"])

}

func Test_HttpContentLoader_Load(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	server := testServer("the body", time.Millisecond*0)
	defer server.Close()

	loader := NewHttpContentLoader()
	mockParser := NewMockContentParser(ctrl)
	mockParser.EXPECT().Parse(gomock.Any(), gomock.Any()).
		Do(func(c *MemoryContent, in io.Reader) {
			body, err := ioutil.ReadAll(in)
			a.NoError(err)
			a.Equal("the body", string(body))
			c.head = StringFragment("some head content")
		})

	loader.parser["text/html"] = mockParser

	c, err := loader.Load(NewFetchDefinition(server.URL))
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	a.Equal(server.URL, c.URL())
	eqFragment(t, "some head content", c.Head())
	a.Equal(0, len(c.Body()))
}

func Test_HttpContentLoader_LoadStream(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	loader := &HttpContentLoader{}
	c, err := loader.Load(NewFetchDefinition(server.URL))
	a.NoError(err)

	a.NotNil(c.Reader())
	body, err := ioutil.ReadAll(c.Reader())
	a.NoError(err)
	a.Equal("{}", string(body))
}

func Test_HttpContentLoader_LoadError500(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", 500)
	}))
	defer server.Close()

	loader := &HttpContentLoader{}
	c, err := loader.Load(NewFetchDefinition(server.URL))
	a.Error(err)
	a.Nil(c)
	a.Contains(err.Error(), "http 500")
}

func Test_HttpContentLoader_LoadErrorNetwork(t *testing.T) {
	a := assert.New(t)

	loader := &HttpContentLoader{}
	c, err := loader.Load(NewFetchDefinition("..."))
	a.Error(err)
	a.Nil(c)
	a.Contains(err.Error(), "unsupported protocol scheme")
}
