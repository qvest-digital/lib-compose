package util

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_CacheStrategy_readCookieValue(t *testing.T) {
	a := assert.New(t)

	v, found := ReadCookieValue(http.Header{"Cookie": {"foo=bar"}}, "foo")
	a.True(found)
	a.Equal("bar", v)

	v, found = ReadCookieValue(http.Header{"Cookie": {`foo="bar"`}}, "foo")
	a.True(found)
	a.Equal("bar", v)

	v, found = ReadCookieValue(http.Header{"Cookie": {"foo"}}, "foo")
	a.True(found)
	a.Equal("", v)

	v, found = ReadCookieValue(http.Header{"Cookie": {";"}}, "foo")
	a.False(found)

	v, found = ReadCookieValue(http.Header{"Cookie": {""}}, "foo")
	a.False(found)

	v, found = ReadCookieValue(http.Header{}, "foo")
	a.False(found)
}
