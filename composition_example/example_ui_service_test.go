package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/yosssi/gohtml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_integration_test(t *testing.T) {
	a := assert.New(t)

	s := httptest.NewServer(handler())
	defer s.Close()

	r, err := http.Get(s.URL)
	a.NoError(err)
	a.Equal(200, r.StatusCode)

	body, err := ioutil.ReadAll(r.Body)
	a.NoError(err)

	expected, err := ioutil.ReadFile("./expected_test_result.html")
	expectedS := strings.Replace(string(expected), "http://127.0.0.1:8080", s.URL, -1)

	a.NoError(err)
	htmlEqual(t, expectedS, string(body))
}

func htmlEqual(t *testing.T, expected, actual string) {
	assert.Equal(t, gohtml.Format(expected), gohtml.Format(actual))
}
