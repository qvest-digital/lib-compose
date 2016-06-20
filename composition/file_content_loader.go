package composition

import (
	"errors"
	"fmt"
	"github.com/tarent/lib-compose/logging"
	"os"
	"strings"
	"time"
)

var ResponseProcessorsNotApplicable = errors.New("request processors are not apliable on file content")

type FileContentLoader struct {
	parser ContentParser
}

func NewFileContentLoader() *FileContentLoader {
	return &FileContentLoader{
		parser: &HtmlContentParser{},
	}
}

func (loader *FileContentLoader) Load(fd *FetchDefinition) (Content, error, int) {
	if fd.RespProc != nil {
		return nil, ResponseProcessorsNotApplicable, 502
	}
	filename := strings.TrimPrefix(fd.URL, FileURLPrefix)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %v: %v", fd.URL, err), 502
	}

	c := NewMemoryContent()
	c.url = fd.URL
	c.httpStatusCode = 200

	if strings.HasSuffix(filename, ".html") {
		parsingStart := time.Now()
		err := loader.parser.Parse(c, f)
		logging.Logger.
			WithField("full_url", c.URL()).
			WithField("duration", time.Since(parsingStart)).
			Debug("content parsing")
		f.Close()
		return c, err, c.httpStatusCode
	}

	c.reader = f
	return c, nil, c.httpStatusCode
}
