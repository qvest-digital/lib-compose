package composition

import (
	"github.com/tarent/lib-compose/cache"
	"github.com/tarent/lib-compose/logging"
	"strings"
	"time"
)

type CachingContentLoader struct {
	httpContentLoader ContentLoader
	fileContentLoader ContentLoader
	cache             Cache
}

func NewCachingContentLoader() *CachingContentLoader {
	c := cache.NewCache(1000, 50*1024*1024, time.Minute*20)
	c.LogEvery(time.Second * 5)
	return &CachingContentLoader{
		httpContentLoader: NewHttpContentLoader(),
		fileContentLoader: NewFileContentLoader(),
		cache:             c,
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
