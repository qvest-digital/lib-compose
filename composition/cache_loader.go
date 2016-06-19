package composition

import (
	"github.com/tarent/lib-compose/cache"
	"strings"
)

type CachingContentLoader struct {
	httpContentLoader ContentLoader
	fileContentLoader ContentLoader
	cache             cache.Cache
}

func NewCachingContentLoader() *CachingContentLoader {
	return &CachingContentLoader{
		httpContentLoader: NewHttpContentLoader(),
		fileContentLoader: NewFileContentLoader(),
	}
}

func (loader *CachingContentLoader) Load(fd *FetchDefinition) (Content, error) {
	hash := fd.Hash()

	if c, exist := loader.cache.Get(hash); exist {
		return c.(Content), nil
	}

	c, err := loader.load(fd)
	if err == nil {
		if fd.IsCachable(c.HttpHeader()) {
			loader.cache.Set(hash, c.MemorySize(), c)
		}
	}
	return c, err
}

func (loader *CachingContentLoader) load(fd *FetchDefinition) (Content, error) {
	if strings.HasPrefix(fd.URL, FileURLPrefix) {
		return loader.fileContentLoader.Load(fd)
	}
	return loader.httpContentLoader.Load(fd)
}
