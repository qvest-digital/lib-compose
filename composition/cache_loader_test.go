package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func Test_CacheLoader_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	// given:
	fd := NewFetchDefinition("/foo")
	c := NewMemoryContent()
	c.name = "/foo"

	// and a cache returning the memory content by the hash
	cacheMocK := NewMockCache(ctrl)
	cacheMocK.EXPECT().Get(fd.Hash()).Return(c, true)

	// when: we load the object
	loader := NewCachingContentLoader(cacheMocK, true, true)

	// it is returned
	result, err := loader.Load(fd)
	a.NoError(err)
	a.Equal(c, result)
}

func Test_CacheLoader_No_Cache_Lookup_For_Uncachable_Objects(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	// given a uncachable fetchdefinition (no cache strategy):
	fd := NewFetchDefinition("/foo")
	fd.CacheStrategy = nil

	// and a mocked cache
	cacheMocK := NewMockCache(ctrl)

	// and mocked Content Loaders
	fileContentLoaderMock := NewMockContentLoader(ctrl)
	httpContentLoaderMock := NewMockContentLoader(ctrl)
	loader := NewCachingContentLoader(cacheMocK, true, true)
	loader.fileContentLoader = fileContentLoaderMock
	loader.httpContentLoader = httpContentLoaderMock

	// and a mocked Content
	contentMock := NewMockContent(ctrl)

	// and a cache returning the memory content by the hash
	fileContentLoaderMock.EXPECT().Load(gomock.Any()).Times(0)
	httpContentLoaderMock.EXPECT().Load(fd).Times(1).Return(contentMock, nil)
	contentMock.EXPECT().HttpHeader().Times(1).Return(nil)
	contentMock.EXPECT().HttpStatusCode().Times(1).Return(0)
	cacheMocK.EXPECT().Get(fd.Hash()).Times(0)

	// when: we load the object

	// it is returned
	_, err := loader.Load(fd)
	a.NoError(err)
}

func Test_CacheLoader_NoLookupForPostRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	// given:
	fd := NewFetchDefinition("/foo")
	fd.Method = "POST"
	c := NewMemoryContent()
	c.name = "/foo"

	// and a cache returning nothing
	cacheMocK := NewMockCache(ctrl)
	httpLoaderMocK := NewMockContentLoader(ctrl)
	httpLoaderMocK.EXPECT().Load(gomock.Any()).Return(c, nil)

	// when: we load the object
	loader := NewCachingContentLoader(cacheMocK, true, true)
	loader.httpContentLoader = httpLoaderMocK

	// it is returned
	result, err := loader.Load(fd)
	a.NoError(err)
	a.Equal(c, result)
}

func Test_CacheLoader_NotFound(t *testing.T) {
	tests := []struct {
		url      string
		method   string
		cachable bool
	}{
		{"http://example.de", "GET", true},
		{"file:///some/file", "GET", true},
	}
	for _, test := range tests {
		ctrl := gomock.NewController(t)
		a := assert.New(t)

		// given:
		c := NewMemoryContent()
		c.name = test.url
		c.httpStatusCode = 200
		fd := NewFetchDefinition(test.url)
		fd.Method = test.method

		// and a cache returning nothing
		cacheMocK := NewMockCache(ctrl)
		cacheMocK.EXPECT().Get(gomock.Any()).Return(nil, false)
		if test.cachable {
			cacheMocK.EXPECT().Set(fd.Hash(), fd.URL, c.MemorySize(), c)
		}
		// and a loader delegating to
		loaderMock := NewMockContentLoader(ctrl)
		loaderMock.EXPECT().Load(gomock.Any()).Return(c, nil)

		// when: we load the object
		loader := NewCachingContentLoader(cacheMocK, true, true)
		if test.url == "file:///some/file" {
			loader.fileContentLoader = loaderMock
		} else {
			loader.httpContentLoader = loaderMock
		}

		// it is returned
		result, err := loader.Load(fd)
		a.NoError(err)
		a.Equal(c, result)
		ctrl.Finish()
	}
}

func Test_CacheLoader_NotFound_With_Stream(t *testing.T) {
	tests := []struct {
		url      string
		method   string
		reader   io.ReadCloser
		cachable bool
	}{
		{"http://example.de", "GET", ioutil.NopCloser(strings.NewReader("foobar")), true},
		{"file:///some/file", "GET", ioutil.NopCloser(strings.NewReader("foobar")), true},
	}
	for _, test := range tests {
		ctrl := gomock.NewController(t)
		a := assert.New(t)

		// given:
		c := NewMemoryContent()
		c.name = test.url
		c.httpStatusCode = 200
		c.reader = test.reader
		fd := NewFetchDefinition(test.url)
		fd.Method = test.method

		// and a cache returning nothing
		cacheMocK := NewMockCache(ctrl)
		cacheMocK.EXPECT().Get(gomock.Any()).Return(nil, false)
		if test.cachable {
			cacheMocK.EXPECT().Set(fd.Hash(), fd.URL, c.MemorySize(), CWMatcher{})
		}
		// and a loader delegating to
		loaderMock := NewMockContentLoader(ctrl)
		loaderMock.EXPECT().Load(gomock.Any()).Return(c, nil)

		// when: we load the object
		loader := NewCachingContentLoader(cacheMocK, true, true)
		if test.url == "file:///some/file" {
			loader.fileContentLoader = loaderMock
		} else {
			loader.httpContentLoader = loaderMock
		}

		// it is returned
		result, err := loader.Load(fd)
		resultbytes, err := ioutil.ReadAll(result.Reader())
		resultstring := string(resultbytes)
		a.NoError(err)
		a.Equal("foobar", resultstring)
		ctrl.Finish()
	}
}

func Test_Content_Wrapper_Reader(t *testing.T) {
	//given
	toTest := &ContentWrapper{streamBytes: []byte("foobar")}

	//when
	result, err := ioutil.ReadAll(toTest.Reader())
	resultStr := string(result)

	//then
	assert.NoError(t, err)
	assert.Equal(t, "foobar", resultStr)

}

type CWMatcher struct {
}

//Checks if a given object is a ContentWrapper
func (CWMatcher) Matches(cw interface{}) bool {
	if reflect.TypeOf(cw) == reflect.TypeOf(&ContentWrapper{}) {
		return true
	}
	return false
}
func (CWMatcher) String() string {
	return "is a ContentWrapper"
}
