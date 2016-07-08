package composition

import (
	"bytes"
	"io"
	"io/ioutil"
	"lib-compose/logging"
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
			if c.Reader() != nil {
				var streamBytes []byte
				streamBytes, err = ioutil.ReadAll(c.Reader())
				cw := &ContentWrapper{
					Content:     c,
					streamBytes: streamBytes,
				}
				loader.cache.Set(hash, fd.URL, c.MemorySize(), cw)
				c.SetReader(cw.Reader())
			} else {
				loader.cache.Set(hash, fd.URL, c.MemorySize(), c)
			}
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

type ContentWrapper struct {
	Content
	streamBytes []byte
}

func (cw *ContentWrapper) Reader() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(cw.streamBytes))
}
