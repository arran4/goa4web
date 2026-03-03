package httputil

import (
	"bytes"
	"net/http"
	"time"
)

// RangeResponseWriter is an http.ResponseWriter that buffers the response and serves it
// using http.ServeContent, which natively supports HTTP Range requests.
type RangeResponseWriter struct {
	w          http.ResponseWriter
	req        *http.Request
	buf        *bytes.Buffer
	statusCode int
	headers    http.Header
}

// NewRangeResponseWriter creates a new RangeResponseWriter.
func NewRangeResponseWriter(w http.ResponseWriter, req *http.Request) *RangeResponseWriter {
	return &RangeResponseWriter{
		w:       w,
		req:     req,
		buf:     &bytes.Buffer{},
		headers: make(http.Header),
	}
}

// Header implements http.ResponseWriter.
func (rw *RangeResponseWriter) Header() http.Header {
	return rw.headers
}

// Write implements http.ResponseWriter.
func (rw *RangeResponseWriter) Write(b []byte) (int, error) {
	return rw.buf.Write(b)
}

// WriteHeader implements http.ResponseWriter.
func (rw *RangeResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// Serve flushes the buffered response to the underlying http.ResponseWriter
// using http.ServeContent, enabling Range request support.
func (rw *RangeResponseWriter) Serve() {
	if rw.statusCode != 0 && rw.statusCode != http.StatusOK {
		// If it's an error or redirect, don't use ServeContent, just copy headers and body.
		for k, v := range rw.headers {
			for _, val := range v {
				rw.w.Header().Add(k, val)
			}
		}
		rw.w.WriteHeader(rw.statusCode)
		rw.w.Write(rw.buf.Bytes())
		return
	}

	// For 200 OK responses, use ServeContent.
	for k, v := range rw.headers {
		for _, val := range v {
			rw.w.Header().Add(k, val)
		}
	}

	reader := bytes.NewReader(rw.buf.Bytes())
	// ServeContent handles Range header processing correctly when passed a zero time.Time.
	http.ServeContent(rw.w, rw.req, "", time.Time{}, reader)
}
