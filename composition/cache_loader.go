package composition

import (
	"bytes"
	"github.com/tarent/lib-compose/logging"
	"io"
	"io/ioutil"
	"strings"
)

type CachingContentLoader struct {
	httpContentLoader ContentLoader
	fileContentLoader ContentLoader
	cache             Cache
}

func NewCachingContentLoader(cache Cache, collectLinks bool, collectScripts bool) *CachingContentLoader {
	return &CachingContentLoader{
		httpContentLoader: NewHttpContentLoader(collectLinks, collectScripts),
		fileContentLoader: NewFileContentLoader(collectLinks, collectScripts),
		cache:             cache,
	}
}

func (loader *CachingContentLoader) Load(fd *FetchDefinition) (Content, error) {
	hash := fd.Hash()

	if fd.Method == "GET" && fd.IsReadableFromCache() {
		if cFromCache, exist := loader.cache.Get(hash); exist {
			logging.Cacheinfo(fd.URL, true)
			return cFromCache.(Content), nil
		}
	}
	logging.Cacheinfo(fd.URL, false)
	c, err := loader.load(fd)
	if err == nil {
		if fd.IsCacheable(c.HttpStatusCode(), c.HttpHeader()) {
			if c.Reader() != nil {
				var streamBytes []byte
				streamBytes, err = ioutil.ReadAll(c.Reader())
				if err == nil {
					cw := &ContentWrapper{
						Content:     c,
						streamBytes: streamBytes,
					}
					loader.cache.Set(hash, fd.URL, c.MemorySize(), cw)
					return cw, nil
				}
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
