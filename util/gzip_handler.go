package util

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	headerVary            = "Vary"
	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentType     = "Content-Type"
	headerContentLength   = "Content-Length"
	encodingGzip          = "gzip"
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
	New: func() interface{} { w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed); return w },
}

// Transparently gzip the response body if
// the client supports it (via the Accept-Encoding header)
// and the response Content-type starts with one of GzipCompressableTypes.
//
// In difference to the most implementations found in the web,
// we do the decision of compression in the Writer, when the Content-Type is determined.
func NewGzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(headerVary, headerAcceptEncoding)

		if acceptsGzip(r) {
			gzWriter := NewGzipResponseWriter(w)
			defer gzWriter.Close()
			h.ServeHTTP(gzWriter, r)
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
			grw.Header().Del(headerContentLength)
			grw.Header().Set(headerContentEncoding, encodingGzip)
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
		if _, ok := grw.Header()[headerContentType]; !ok {
			// Set content-type if not present. Otherwise golang would make application/gzip out of that.
			grw.Header().Set(headerContentType, http.DetectContentType(b))
		}
		grw.WriteHeader(http.StatusOK)
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
	if header.Get(headerContentEncoding) != "" {
		return false
	}

	// check if we should compress for this content type
	ct := header.Get(headerContentType)
	for _, t := range GzipCompressableTypes {
		if strings.HasPrefix(ct, t) {
			return true
		}
	}
	return false
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get(headerAcceptEncoding), encodingGzip)
}
