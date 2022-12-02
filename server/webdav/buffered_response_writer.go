package webdav

import (
	"net/http"
)

type bufferedResponseWriter struct {
	statusCode int
	data       []byte
	header     http.Header
}

func (w *bufferedResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *bufferedResponseWriter) Write(bytes []byte) (int, error) {
	w.data = append(w.data, bytes...)
	return len(bytes), nil
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode == 0 {
		w.statusCode = statusCode
	}
}

func (w *bufferedResponseWriter) WriteToResponse(rw http.ResponseWriter) (int, error) {
	h := rw.Header()
	for k, vs := range w.header {
		for _, v := range vs {
			h.Add(k, v)
		}
	}
	rw.WriteHeader(w.statusCode)
	return rw.Write(w.data)
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		statusCode: 0,
	}
}
