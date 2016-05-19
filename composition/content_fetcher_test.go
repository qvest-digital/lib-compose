package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_ContentFetcher_FetchDefinitionHash(t *testing.T) {
	a := assert.New(t)
	tests := []struct {
		fd1 *FetchDefinition
		fd2 *FetchDefinition
		eq  bool
	}{
		{
			NewFetchDefinition("/foo"),
			NewFetchDefinition("/foo"),
			true,
		},
		{
			NewFetchDefinition("/foo"),
			NewFetchDefinition("/bar"),
			false,
		},
		{
			&FetchDefinition{
				URL:      "/foo",
				Timeout:  time.Second,
				Header:   http.Header{"Some": {"header"}},
				Required: false,
			},
			&FetchDefinition{
				URL:      "/foo",
				Timeout:  time.Second * 42,
				Header:   http.Header{"Some": {"header"}},
				Required: true,
			},
			true,
		},
		{
			&FetchDefinition{
				URL:    "/foo",
				Header: http.Header{"Some": {"header"}},
			},
			&FetchDefinition{
				URL:    "/foo",
				Header: http.Header{"Some": {"other header"}},
			},
			false,
		},
	}

	for _, t := range tests {
		if t.eq {
			a.Equal(t.fd1.Hash(), t.fd2.Hash())
		} else {
			a.NotEqual(t.fd1.Hash(), t.fd2.Hash())
		}
	}
}

func Test_ContentFetcher_FetchingWithDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	loader := NewMockContentLoader(ctrl)
	barFd := getFetchDefinitionMock(ctrl, loader, "/bar", nil, time.Millisecond*2, map[string]interface{}{"foo": "bar"})
	fooFd := getFetchDefinitionMock(ctrl, loader, "/foo", []*FetchDefinition{barFd}, time.Millisecond*2, map[string]interface{}{"bli": "bla"})
	bazzFd := getFetchDefinitionMock(ctrl, loader, "/bazz", []*FetchDefinition{barFd}, time.Millisecond, map[string]interface{}{})

	fetcher := NewContentFetcher(nil)
	fetcher.contentLoader = loader

	fetcher.AddFetchJob(fooFd)
	fetcher.AddFetchJob(bazzFd)

	results := fetcher.WaitForResults()

	a.Equal(3, len(results))

	a.Equal("/foo", results[0].Def.URL)
	a.Equal("/bazz", results[1].Def.URL)
	a.Equal("/bar", results[2].Def.URL)

	meta := fetcher.MetaJSON()
	a.Equal("bar", meta["foo"])
	a.Equal("bla", meta["bli"])
}

func Test_ContentFetcher_NewFetchDefinitionFromRequest(t *testing.T) {
	a := assert.New(t)

	reader := strings.NewReader("the body")
	r, err := http.NewRequest("POST", "https://example.com/content?foo=bar", reader)
	a.NoError(err)

	r.Header = http.Header{"Content-Type": {"text/html"}}

	fd := NewFetchDefinitionFromRequest("http://upstream:8080/", r)
	a.Equal("http://upstream:8080/content?foo=bar", fd.URL)
	a.Equal(10*time.Second, fd.Timeout)
	a.Equal(true, fd.Required)
	a.Equal("text/html", fd.Header.Get("Content-Type"))
	a.Equal("POST", fd.Method)
	b, err := ioutil.ReadAll(fd.Body)
	a.NoError(err)
	a.Equal("the body", string(b))
}

func getFetchDefinitionMock(ctrl *gomock.Controller, loaderMock *MockContentLoader, url string, requiredContent []*FetchDefinition, loaderBlocking time.Duration, metaJSON map[string]interface{}) *FetchDefinition {
	fd := NewFetchDefinition(url)
	fd.Timeout = time.Second * 42

	content := NewMockContent(ctrl)
	content.EXPECT().
		RequiredContent().
		Return(requiredContent)

	content.EXPECT().
		Meta().
		Return(metaJSON)

	loaderMock.EXPECT().
		Load(fd).
		Do(
			func(fetchDefinition *FetchDefinition) {
				time.Sleep(loaderBlocking)
			}).
		Return(content, nil)

	return fd
}
