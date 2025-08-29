package response

import (
	"fmt"
	"io"

	"github.com/devasherr/tcp-http/internal/headers"
)

type Response struct{}

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func (w *Writer) WriteHeaders(headers *headers.Headers) error {
	h := []byte{}
	headers.Foreach(func(n, v string) {
		h = fmt.Appendf(h, "%s: %s\r\n", n, v)
	})

	h = fmt.Append(h, "\r\n")
	_, err := w.writer.Write(h)
	return err
}

func (w *Writer) WriteStatusLine(statStatusCode StatusCode) error {
	statusLine := []byte{}
	switch statStatusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("unknown error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	// check for error maybe ??
	return n, err
}
