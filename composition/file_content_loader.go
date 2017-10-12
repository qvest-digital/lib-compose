package composition

import (
	"errors"
	"fmt"
	"github.com/tarent/go-log-middleware/logging"
	"os"
	"path/filepath"
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

func (loader *FileContentLoader) Load(fd *FetchDefinition) (Content, error) {
	if fd.RespProc != nil {
		return nil, ResponseProcessorsNotApplicable
	}

	path := strings.TrimPrefix(fd.URL, FileURLPrefix)
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		path = filepath.Join(path, "index.html")
	} else if os.IsNotExist(err) {
		c := NewMemoryContent()
		c.name = fd.Name
		c.httpStatusCode = 404
		return c, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %v: %v", fd.URL, err)
	}

	c := NewMemoryContent()
	c.name = fd.Name
	c.httpStatusCode = 200

	if strings.HasSuffix(path, ".html") {
		parsingStart := time.Now()
		err := loader.parser.Parse(c, f)
		logging.Logger.
			WithField("full_url", fd.URL).
			WithField("duration", time.Since(parsingStart)).
			Debug("content parsing")
		f.Close()
		return c, err
	}

	c.reader = f
	return c, nil
}
