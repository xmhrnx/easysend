package webshare

import (
	"io"
	"net/http"
)

type countingWriter struct {
	http.ResponseWriter
	fn func(int64)
}

type nopRW struct{ io.Writer }

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.ResponseWriter.Write(p)
	if c.fn != nil && n > 0 {
		c.fn(int64(n))
	}
	return n, err
}

func (nopRW) Header() http.Header { return http.Header{} }
func (nopRW) WriteHeader(int)     {}
