package composition

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_FetchDefinition_NewFetchDefinitionFromRequest(t *testing.T) {
	a := assert.New(t)

	reader := strings.NewReader("the body")
	r, err := http.NewRequest("POST", "https://example.com/content?foo=bar", reader)
	a.NoError(err)

	r.Header = http.Header{
		"Content-Type":    {"text/html"},
		"Cookie":          {"aa=bb;"},
		"Accept-Encoding": {"gzip"}, // should not be copied
	}

	fd := NewFetchDefinitionFromRequest("http://upstream:8080/", r)
	a.Equal("http://upstream:8080/content?foo=bar", fd.URL)
	a.Equal(10*time.Second, fd.Timeout)
	a.Equal(true, fd.Required)

	a.Equal("text/html", fd.Header.Get("Content-Type"))
	a.Equal("aa=bb;", fd.Header.Get("Cookie"))
	a.Equal("", fd.Header.Get("Accept-Encoding"))

	a.Equal("POST", fd.Method)
	b, err := ioutil.ReadAll(fd.Body)
	a.NoError(err)
	a.Equal("the body", string(b))
}

func Test_FetchDefinition_use_DefaultErrorHandler_if_not_set(t *testing.T) {
	a := assert.New(t)

	fd := NewFetchDefinitionWithErrorHandler("http://upstream:8080/", nil)
        a.Equal(NewDefaultErrorHandler(), fd.ErrHandler)
}
