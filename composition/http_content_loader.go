package composition

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/tarent/lib-compose/logging"
)

type HttpContentLoader struct {
	parser map[string]ContentParser
}

func NewHttpContentLoader() *HttpContentLoader {
	return &HttpContentLoader{
		parser: map[string]ContentParser{
			"text/html": &HtmlContentParser{},
		},
	}
}

// TODO: Should we filter the headers, which we forward here, or is it correct to copy all of them?
func (loader *HttpContentLoader) Load(fd *FetchDefinition) (Content, error, int) {
	client := &http.Client{Timeout: fd.Timeout}
	statusCode:=502

	var err error
	request, err := http.NewRequest(fd.Method, fd.URL, fd.Body)
	if err != nil {
		return nil, err, 0
	}
	request.Header = fd.Header
	if request.Header == nil {
		request.Header = http.Header{}
	}
	request.Header.Set("User-Agent", "lib-compose")

	start := time.Now()

	resp, err := client.Do(request)

	if resp != nil {
		statusCode = resp.StatusCode
	}


	logging.Call(request, resp, start, err)
	if err != nil {
		return nil, err, statusCode
	}

	if fd.RespProc != nil {
		err = fd.RespProc.Process(resp, fd.URL)
	}
	if err != nil {
		return nil, err, statusCode
	}

	if statusCode < 200 || statusCode > 399 {
		return nil, fmt.Errorf("(http %v) on loading url %q", statusCode, fd.URL), statusCode
	}

	c := NewMemoryContent()
	c.url = fd.URL
	c.httpHeader = resp.Header
	c.httpStatusCode = statusCode

	// take the first parser for the content type
	// direct access to the map does not work, because the
	// content type may have encoding information at the end
	reponseType := resp.Header.Get("Content-Type")
	responseNoCompositionHeader := resp.Header.Get("X-No-Composition")
	if responseNoCompositionHeader == "" {
		for contentType, parser := range loader.parser {
			if strings.HasPrefix(reponseType, contentType) {
				defer func() {
					// read and close the body, to make reuse of tcp connections
					ioutil.ReadAll(resp.Body)
					resp.Body.Close()
				}()
				parsingStart := time.Now()
				err := parser.Parse(c, resp.Body)
				logging.Logger.
					WithField("full_url", c.URL()).
					WithField("duration", time.Since(parsingStart)).
					Debug("content parsing")
				return c, err, statusCode
			}
		}
	}

	c.reader = resp.Body
	return c, nil, statusCode
}
