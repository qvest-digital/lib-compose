package composition

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/lib-servicediscovery/servicediscovery"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	loader := NewHttpContentLoader(true, true)
	mockParser := NewMockContentParser(ctrl)
	mockParser.EXPECT().Parse(gomock.Any(), gomock.Any()).
		Do(func(c *MemoryContent, in io.Reader) {
			body, err := ioutil.ReadAll(in)
			a.NoError(err)
			a.Equal("the body", string(body))
			c.head = NewStringFragment("some head content")
		})

	loader.parser["text/html"] = mockParser

	fd := NewFetchDefinition(server.URL)
	fd.Name = "content"
	c, err := loader.Load(fd)
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	a.Equal("content", c.Name())
	eqFragment(t, "some head content", c.Head())
	a.Equal(0, len(c.Body()))
}

func Test_HttpContentLoader_Load_ResponseProcessor(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)
	request := &http.Request{}
	request.URL = &url.URL{}

	server := testServer("the body", time.Millisecond*0)
	defer server.Close()

	loader := NewHttpContentLoader(true, true)
	mockParser := NewMockContentParser(ctrl)
	mockParser.EXPECT().Parse(gomock.Any(), gomock.Any()).
		Do(func(c *MemoryContent, in io.Reader) {
			body, err := ioutil.ReadAll(in)
			a.NoError(err)
			a.Equal("the body", string(body))
			c.head = NewStringFragment("some head content")
		})

	loader.parser["text/html"] = mockParser

	mockResponseProcessor := NewMockResponseProcessor(ctrl)
	mockResponseProcessor.EXPECT().Process(gomock.Any(), gomock.Any())
	c, err := loader.Load(NewFetchDefinition(server.URL).WithResponseProcessor(mockResponseProcessor).FromRequest(request))
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
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

	loader := NewHttpContentLoader(true, true)
	mockParser := NewMockContentParser(ctrl)
	mockParser.EXPECT().Parse(gomock.Any(), gomock.Any()).
		Do(func(c *MemoryContent, in io.Reader) {
			body, err := ioutil.ReadAll(in)
			a.NoError(err)
			a.Equal("the body", string(body))
			c.head = NewStringFragment("some head content")
		})

	loader.parser["text/html"] = mockParser

	fd := NewFetchDefinition(server.URL)
	fd.Header = http.Header{"X-Foo": {"bar"}}
	fd.Method = "POST"
	fd.Body = strings.NewReader("post content")

	c, err := loader.Load(fd)
	a.NoError(err)
	a.NotNil(c)
	a.Nil(c.Reader())
	a.Equal(server.URL, c.Name())
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

func Test_HttpContentLoader_LoadStream_No_Composition_Header(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-No-Composition", "1")
		w.Header().Set("Content-Type", "text/html")
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

func Test_HttpContentLoader_Pass_404(t *testing.T) {
	a := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("{}"))
	}))

	defer server.Close()

	loader := &HttpContentLoader{}
	c, err := loader.Load(NewFetchDefinition(server.URL))
	a.Error(err)
	a.Equal(404, c.HttpStatusCode())
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
	a.Contains(err.Error(), "http 500")
	assert.True(t, c.HttpStatusCode() == 500)
}

func Test_HttpContentLoader_LoadErrorNetwork(t *testing.T) {
	a := assert.New(t)

	loader := &HttpContentLoader{}
	_, err := loader.Load(NewFetchDefinition("..."))
	a.Error(err)
	a.Contains(err.Error(), "unsupported protocol scheme")
}

func Test_HttpContentLoader_FollowRedirects(t *testing.T) {
	a := assert.New(t)

	for _, status := range []int{301, 302} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/redirected" {
				http.Redirect(w, r, "/redirected", status)
				return
			}
			w.Write([]byte("ok"))
		}))

		loader := &HttpContentLoader{}
		fd := NewFetchDefinition(server.URL)
		fd.FollowRedirects = true
		c, err := loader.Load(fd)
		a.NoError(err)
		a.Equal(200, c.HttpStatusCode())

		a.NotNil(c.Reader())
		body, err := ioutil.ReadAll(c.Reader())
		a.NoError(err)
		a.Equal("ok", string(body))

		server.Close()
	}
}

func Test_HttpContentLoader_DoNotFollowRedirects(t *testing.T) {
	a := assert.New(t)

	for _, status := range []int{301, 302, 303, 305, 307, 308} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/redirected" {
				http.Redirect(w, r, "/redirected", status)
				return
			}
			w.Write([]byte("ok"))
		}))

		loader := &HttpContentLoader{}
		fd := NewFetchDefinition(server.URL)
		fd.FollowRedirects = false
		c, err := loader.Load(fd)
		a.NoError(err)

		a.Equal(status, c.HttpStatusCode())
		a.Equal("/redirected", c.HttpHeader().Get("Location"))

		server.Close()
	}
}

func Test_HttpContentLoader_DiscoverServiceInUrl(t *testing.T) {

	a := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// given
	loader := &HttpContentLoader{}

	mockServiceDiscovery := servicediscovery.NewMockServiceDiscovery(ctrl)
	mockServiceDiscovery.EXPECT().DiscoverService("serviceName").Return("10.0.0.1", "42", nil)

	// when
	url, _ := loader.discoverServiceInUrl("http://serviceName/test.jpg", mockServiceDiscovery)

	// then
	a.Equal(url, "http://10.0.0.1:42/test.jpg")
}

func Test_HttpContentLoader_DiscoverServiceInUrlRawIp(t *testing.T) {

	a := assert.New(t)

	cases := [][]string{
		{"http://127.0.0.1:80/test.jpg", "http://127.0.0.1:80/test.jpg"},
		{"http://127.0.0.1/test.jpg", "http://127.0.0.1/test.jpg"},
	}

	for _, v := range cases {

		fmt.Println(v)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// given
		loader := &HttpContentLoader{}

		mockServiceDiscovery := servicediscovery.NewMockServiceDiscovery(ctrl)

		// when
		url, _ := loader.discoverServiceInUrl(v[0], mockServiceDiscovery)

		// then
		a.Equal(url, v[1])
	}

}

func Test_HttpContentLoader_DiscoverServiceInUrlWithPortError(t *testing.T) {

	a := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// given
	loader := &HttpContentLoader{}

	mockServiceDiscovery := servicediscovery.NewMockServiceDiscovery(ctrl)

	// when
	url, err := loader.discoverServiceInUrl("http://serviceName:80/test.jpg", mockServiceDiscovery)

	// then
	a.Equal(url, "")
	a.EqualError(err, "Service name with port given, this is not allowed. The port will be resolved by service discovery!")

}

func testServer(content string, timeout time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		time.Sleep(timeout)
		w.Write([]byte(content))
	}))
}
