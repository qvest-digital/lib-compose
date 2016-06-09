package util

import (
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func Test_GzipHandler_CompressableType(t *testing.T) {
	a := assert.New(t)

	// given: a server which returns text
	server := httptest.NewServer(NewGzipHandler(test_text_handler()))

	// when: we call the server with gzip accept header
	r, err := http.NewRequest("GET", server.URL, nil)
	a.NoError(err)
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	a.NoError(err)

	// then: it is gzipped
	a.Equal("text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	a.Equal("gzip", resp.Header.Get("Content-Encoding"))

	gzBytes, err := ioutil.ReadAll(resp.Body)
	a.NoError(err)
	a.Equal(strconv.Itoa(len(gzBytes)), resp.Header.Get("Content-Length"))

	// and: we can uncomress it
	reader, err := gzip.NewReader(bytes.NewBuffer(gzBytes))
	a.NoError(err)
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	a.NoError(err)

	a.Equal("Hello World", string(bytes))
}

func Test_GzipHandler_NotCompressingTwice(t *testing.T) {
	a := assert.New(t)

	// given: a server which returns an already compressed stream
	server := httptest.NewServer(NewGzipHandler(test_already_compressed_handler()))

	// when: we call the server with gzip accept header
	r, err := http.NewRequest("GET", server.URL, nil)
	a.NoError(err)
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	a.NoError(err)

	// then: it is gzipped
	a.Equal("gzip", resp.Header.Get("Content-Encoding"))

	// and: we can uncomress it
	reader, err := gzip.NewReader(resp.Body)
	a.NoError(err)
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	a.NoError(err)

	a.Equal("Hello World", string(bytes))
}

func Test_GzipHandler_CompressableType_NoAccept(t *testing.T) {
	a := assert.New(t)

	// given: a server which returns text
	server := httptest.NewServer(NewGzipHandler(test_text_handler()))

	// when: we call it without gzip accept header
	r, err := http.NewRequest("GET", server.URL, nil)
	a.NoError(err)
	r.Header.Set("Accept-Encoding", "none")

	resp, err := http.DefaultClient.Do(r)
	a.NoError(err)

	// then: it is not encrypted
	a.Equal("", resp.Header.Get("Content-Encoding"))

	bytes, err := ioutil.ReadAll(resp.Body)
	a.NoError(err)

	a.Equal("Hello World", string(bytes))
}

func Test_GzipHandler_NonCompressableType(t *testing.T) {
	a := assert.New(t)

	// given: a server which returns binary data, which is not listed in the content types
	server := httptest.NewServer(NewGzipHandler(test_binary_handler()))

	// when we call it with accept header
	r, err := http.NewRequest("GET", server.URL, nil)
	a.NoError(err)
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	a.NoError(err)

	// then: it is not encrypted
	a.Equal("", resp.Header.Get("Content-Encoding"))

	bytes, err := ioutil.ReadAll(resp.Body)
	a.NoError(err)

	a.Equal([]byte{42}, bytes)
}

func test_text_handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := []byte("Hello World")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)
	})
}

func test_binary_handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpg")
		w.Write([]byte{42})
	})
}

func test_already_compressed_handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gzWriter := gzip.NewWriter(w)
		gzWriter.Write([]byte("Hello World"))
		gzWriter.Close()
	})
}
