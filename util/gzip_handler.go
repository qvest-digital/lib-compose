package util

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
)

var GzipCompressableTypes = []string{
	"text/",
	"application/json",
	"application/xhtml+xml",
	"application/xml",
	"image/svg+xml",
	"application/javascript",
	"application/font-woff",
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} { return gzip.NewWriter(nil) },
}

// Transparently gzip the response body if
// the client supports it (via the Accept-Encoding header)
// and the response Content-type starts with one of GzipCompressableTypes.
//
// In difference to the most implementations found in the web,
// we do the decision of compression in the Writer, when the Content-Type is determined.
func NewGzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(vary, acceptEncoding)

		if acceptsGzip(r) {
			gzWriter := NewGzipResponseWriter(w)
			h.ServeHTTP(gzWriter, r)
			gzWriter.Close()
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type GzipResponseWriter struct {
	writer     io.Writer
	gzipWriter *gzip.Writer
	http.ResponseWriter
}

func NewGzipResponseWriter(w http.ResponseWriter) *GzipResponseWriter {
	return &GzipResponseWriter{ResponseWriter: w}
}

func (grw *GzipResponseWriter) WriteHeader(code int) {
	if grw.writer == nil {
		if isCompressable(grw.Header()) {
			grw.Header().Del("Content-Length")
			grw.Header().Set(contentEncoding, "gzip")
			grw.gzipWriter = gzipWriterPool.Get().(*gzip.Writer)
			grw.gzipWriter.Reset(grw.ResponseWriter)

			grw.writer = grw.gzipWriter
		} else {
			grw.writer = grw.ResponseWriter
		}
	}
	grw.ResponseWriter.WriteHeader(code)
}

func (grw *GzipResponseWriter) Write(b []byte) (int, error) {
	if grw.writer == nil {
		if _, ok := grw.Header()["Content-Type"]; !ok {
			// Set content-type if not present. Otherwise golang would make application/gzip out of that.
			grw.Header().Set("Content-Type", http.DetectContentType(b))
		}
		grw.WriteHeader(200)
	}
	return grw.writer.Write(b)
}

func (grw *GzipResponseWriter) Close() {
	if grw.gzipWriter != nil {
		grw.gzipWriter.Close()
		gzipWriterPool.Put(grw.gzipWriter)
	}
}

func isCompressable(header http.Header) bool {
	// don't compress if it is already encoded
	if header.Get(contentEncoding) != "" {
		return false
	}

	// check if we should compress for this content type
	contentType := header.Get("Content-Type")
	for _, t := range GzipCompressableTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}
	return false
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}
