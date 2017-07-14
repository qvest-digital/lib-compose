package composition

import (
	"errors"
	"fmt"
	"github.com/tarent/lib-compose/logging"
	"github.com/tarent/lib-servicediscovery/servicediscovery"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var redirectAttemptedError = errors.New("do not follow redirects")
var noRedirectFunc = func(req *http.Request, via []*http.Request) error {
	return redirectAttemptedError
}

type HttpContentLoader struct {
	parser map[string]ContentParser
}

func NewHttpContentLoader(collectLinks bool, collectScripts bool) *HttpContentLoader {
	return &HttpContentLoader{
		parser: map[string]ContentParser{
			"text/html": NewHtmlContentParser(collectLinks, collectScripts),
		},
	}
}

// TODO: Should we filter the headers, which we forward here, or is it correct to copy all of them?
func (loader *HttpContentLoader) Load(fd *FetchDefinition) (Content, error) {
	client := &http.Client{Timeout: fd.Timeout}

	c := NewMemoryContent()
	c.name = fd.Name
	c.httpStatusCode = 502

	// redirects can only be stopped by returning an error in the CheckRedirect function
	if !fd.FollowRedirects {
		client.CheckRedirect = noRedirectFunc
	}

	fetchUrl := fd.URL
	if fd.ServiceDiscoveryActive {
		discoveredUrl, err := loader.discoverServiceInUrl(fetchUrl, fd.ServiceDiscovery)
		if err != nil {
			return c, err
		}
		fetchUrl = discoveredUrl
	}

	request, err := http.NewRequest(fd.Method, fetchUrl, fd.Body)
	if err != nil {
		return c, err
	}
	request.Header = fd.Header
	if request.Header == nil {
		request.Header = http.Header{}
	}
	request.Header.Set("User-Agent", "lib-compose")

	start := time.Now()

	resp, err := client.Do(request)
	if resp != nil {
		c.httpStatusCode = resp.StatusCode
		c.httpHeader = resp.Header
	}

	// do not handle our own redirects returns as errors
	if urlError, ok := err.(*url.Error); ok && urlError.Err == redirectAttemptedError {
		logging.Call(request, resp, start, nil)
		return c, nil
	}
	logging.Call(request, resp, start, err)

	if err != nil {
		return c, err
	}

	if fd.RespProc != nil {
		if err := fd.RespProc.Process(resp, fd.URL); err != nil {
			return c, err
		}
	}

	if c.httpStatusCode < 200 || c.httpStatusCode > 399 {
		return c, fmt.Errorf("(http %v) on loading url %q", c.httpStatusCode, fd.URL)
	}

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
					WithField("full_url", fd.URL).
					WithField("duration", time.Since(parsingStart)).
					Debug("content parsing")
				return c, err
			}
		}
	}

	c.reader = resp.Body
	return c, nil
}

func (loader *HttpContentLoader) discoverServiceInUrl(rawUrl string, serviceDiscovery servicediscovery.ServiceDiscovery) (string, error) {

	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	host, origPort, err := net.SplitHostPort(parsedUrl.Host)
	if err != nil {
		if !strings.Contains(err.Error(), "missing port") {
			return "", err
		}
		host = parsedUrl.Host
	}

	if net.ParseIP(host) == nil {
		if origPort != "" {
			return "", fmt.Errorf("Service name with port given, this is not allowed. The port will be resolved by service discovery!")
		}
		ip, port, err := serviceDiscovery.DiscoverService(parsedUrl.Host)
		if err != nil {
			return "", err
		}
		parsedUrl.Host = net.JoinHostPort(ip, port)
	}

	return parsedUrl.String(), nil

}
