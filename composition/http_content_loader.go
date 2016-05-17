package composition

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
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

// TODO: Include Cookies and HTTP Headers from original request to the call
func (loader *HttpContentLoader) Load(fd *FetchDefinition) (Content, error) {
	client := &http.Client{Timeout: fd.Timeout}
	resp, err := client.Get(fd.URL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("(http %v) on loading url %q", resp.StatusCode, fd.URL)
	}

	c := NewMemoryContent()
	c.url = fd.URL
	c.httpHeader = resp.Header

	reponseType := resp.Header.Get("Content-Type")
	for contentType, parser := range loader.parser {
		if strings.HasPrefix(reponseType, contentType) {
			defer func() {
				// read and close the body, to make reuse of tcp connections
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}()

			return c, parser.Parse(c, resp.Body)
		}
	}

	c.reader = resp.Body
	return c, nil
}

func testServer(content string, timeout time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		time.Sleep(timeout)
		w.Write([]byte(content))
	}))
}
