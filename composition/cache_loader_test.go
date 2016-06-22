package composition

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CacheLoader_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	a := assert.New(t)

	// given:
	fd := NewFetchDefinition("/foo")
	c := NewMemoryContent()
	c.url = "/foo"

	// and a cache returning the memory content by the hash
	cacheMocK := NewMockCache(ctrl)
	cacheMocK.EXPECT().Get(fd.Hash()).Return(c, true)

	// when: we load the object
	loader := NewCachingContentLoader()
	loader.cache = cacheMocK

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
		{"http://example.de", "POST", false},
	}
	for _, test := range tests {
		ctrl := gomock.NewController(t)
		a := assert.New(t)

		// given:
		c := NewMemoryContent()
		c.url = test.url
		c.httpStatusCode = 200
		fd := NewFetchDefinition(c.url)
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
		loader := NewCachingContentLoader()
		loader.cache = cacheMocK
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
