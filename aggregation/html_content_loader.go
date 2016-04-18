package aggregation

import (
	"errors"
	"time"
)

type HtmlContentLoader struct {
}

func (loader *HtmlContentLoader) Load(url string, timeout time.Duration) (Content, error) {
	return nil, errors.New("HtmlContentLoader#Load not implemented")
}
