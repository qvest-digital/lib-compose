package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

	c, err, _ := loader.Load(NewFetchDefinition(server.URL))
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	a.Equal(server.URL, c.URL())
	eqFragment(t, "some head content", c.Head())
	a.Equal(0, len(c.Body()))
}

func Test_HttpContentLoader_Load_ResponseProcessor(t *testing.T) {

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

	mockResponseProcessor := NewMockResponseProcessor(ctrl)
	mockResponseProcessor.EXPECT().Process(gomock.Any(), gomock.Any())
	c, err, _ := loader.Load(NewFetchDefinitionWithResponseProcessor(server.URL, mockResponseProcessor))
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	a.Equal(server.URL, c.URL())
	eqFragment(t, "some head content", c.Head())
	a.Equal(0, len(c.Body()))
}

func Test_HttpContentLoader_Load_POST(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		body, err := ioutil.ReadAll(r.Body)
		a.NoError(err)
		a.Equal("post content", string(body))
		a.Equal("POST", r.Method)
		a.Equal("bar", r.Header.Get("X-Foo"))
		w.Write([]byte("the body"))
	}))

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

	fd := NewFetchDefinition(server.URL)
	fd.Header = http.Header{"X-Foo": {"bar"}}
	fd.Method = "POST"
	fd.Body = strings.NewReader("post content")

	c, err, _ := loader.Load(fd)
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
	c, err, _ := loader.Load(NewFetchDefinition(server.URL))
	a.NoError(err)
	a.NotNil(c.Reader())
	body, err := ioutil.ReadAll(c.Reader())
	a.NoError(err)
	a.Equal("{}", string(body))
}

func Test_HttpContentLoader_LoadStream_No_Composition_Header(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-No-Composition", "1")
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	loader := &HttpContentLoader{}
	c, err, _ := loader.Load(NewFetchDefinition(server.URL))
	a.NoError(err)
	a.NotNil(c.Reader())
	body, err := ioutil.ReadAll(c.Reader())
	a.NoError(err)
	a.Equal("{}", string(body))
}

func Test_HttpContentLoader_Pass_404(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("{}"))
	}))

	defer server.Close()

	loader := &HttpContentLoader{}
	c, err, status := loader.Load(NewFetchDefinition(server.URL))
	a.Error(err)
	a.Nil(c)
	a.Equal(404, status)
}

func Test_HttpContentLoader_LoadError500(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", 500)
	}))
	defer server.Close()

	loader := &HttpContentLoader{}
	c, err, statusCode := loader.Load(NewFetchDefinition(server.URL))
	a.Error(err)
	a.Nil(c)
	a.Contains(err.Error(), "http 500")
	assert.True(t, statusCode == 500)
}

func Test_HttpContentLoader_LoadErrorNetwork(t *testing.T) {
	a := assert.New(t)

	loader := &HttpContentLoader{}
	c, err, _ := loader.Load(NewFetchDefinition("..."))
	a.Error(err)
	a.Nil(c)
	a.Contains(err.Error(), "unsupported protocol scheme")
}

func testServer(content string, timeout time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		time.Sleep(timeout)
		w.Write([]byte(content))
	}))
}
