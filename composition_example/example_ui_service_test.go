package main

import (
	"github.com/stretchr/testify/assert"
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
	a.Equal(expectedS, string(body))
}
