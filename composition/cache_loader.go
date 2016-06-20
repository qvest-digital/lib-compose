package composition

import (
	"github.com/tarent/lib-compose/cache"
	"strings"
)

type CachingContentLoader struct {
	httpContentLoader ContentLoader
	fileContentLoader ContentLoader
	cache             *cache.Cache
}

func NewCachingContentLoader() *CachingContentLoader {
	return &CachingContentLoader{
		httpContentLoader: NewHttpContentLoader(),
		fileContentLoader: NewFileContentLoader(),
		cache:             cache.NewCache(10000),
	}
}

func (loader *CachingContentLoader) Load(fd *FetchDefinition) (Content, error) {
	hash := fd.Hash()

	if c, exist := loader.cache.Get(hash); exist {
		println("found: " + fd.URL + " " + hash)
		return c.(Content), nil
	} else {
		println("not found: " + fd.URL + " " + hash)

	}

	c, err := loader.load(fd)
	if err == nil {
		if fd.IsCachable(c.HttpStatusCode(), c.HttpHeader()) {
			println("Set: " + fd.URL + " " + hash)
			loader.cache.Set(hash, c.MemorySize(), c)
		} else {
			println("Not cachable: " + fd.URL + " " + hash)
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
