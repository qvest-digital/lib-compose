package aggregation

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_HtmlContentLoader(t *testing.T) {
	a := assert.New(t)

	loader := &HtmlContentLoader{}
	c, err := loader.Load("/foo", time.Second)
	a.Nil(c)
	a.Error(err)
}
