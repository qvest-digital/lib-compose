package composition

import (
	"github.com/tarent/lib-compose/logging"
	"strings"
)

type CachingContentLoader struct {
	httpContentLoader ContentLoader
	fileContentLoader ContentLoader
	cache             Cache
}

func NewCachingContentLoader(cache Cache) *CachingContentLoader {
	return &CachingContentLoader{
		httpContentLoader: NewHttpContentLoader(),
		fileContentLoader: NewFileContentLoader(),
		cache:             cache,
	}
}

func (loader *CachingContentLoader) Load(fd *FetchDefinition) (Content, error) {
	hash := fd.Hash()

	if cFromCache, exist := loader.cache.Get(hash); exist {
		logging.Cacheinfo(fd.URL, true)
		return cFromCache.(Content), nil
	}
	logging.Cacheinfo(fd.URL, false)
	c, err := loader.load(fd)
	if err == nil {
		if fd.IsCachable(c.HttpStatusCode(), c.HttpHeader()) {
			loader.cache.Set(hash, fd.URL, c.MemorySize(), c)
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
