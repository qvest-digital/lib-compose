package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yosssi/gohtml"
)

func Test_integration_test(t *testing.T) {
	a := assert.New(t)

	s := httptest.NewServer(handler())
	defer s.Close()
	host = s.URL

	r, err := http.Get(s.URL)
	a.NoError(err)
	a.Equal(200, r.StatusCode)

	body, err := ioutil.ReadAll(r.Body)
	a.NoError(err)

	expected, err := ioutil.ReadFile("./expected_test_result.html")
	expectedS := strings.Replace(string(expected), "http://127.0.0.1:8080", s.URL, -1)
	expectedS = removeEmptyLines(expectedS)
	result := removeEmptyLines(string(body))
	ioutil.WriteFile("/tmp/expected", []byte(expectedS), 0644)
	ioutil.WriteFile("/tmp/result", []byte(result), 0644)
	a.NoError(err)
	htmlEqual(t, expectedS, result)
}

func htmlEqual(t *testing.T, expected, actual string) {
	assert.Equal(t, gohtml.Format(expected), gohtml.Format(actual))
}

func removeEmptyLines(stringToProcess string) string {
	re := regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	stringToProcess = re.ReplaceAllString(stringToProcess, "")
	return stringToProcess
}
