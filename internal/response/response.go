package response

import (
	"fmt"
	"io"

	"github.com/devasherr/tcp-http/internal/headers"
)

type Response struct{}

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

func WriteHeaders(w io.Writer, headers *headers.Headers) error {
	var err error = nil
	h := []byte{}
	headers.Foreach(func(n, v string) {
		if err != nil {
			return
		}

		h = fmt.Appendf(h, "%s: %s\r\n", n, v)
	})

	h = fmt.Append(h, "\r\n")
	_, err = w.Write(h)
	return err
}

func WriteStatusLine(w io.Writer, statStatusCode StatusCode) error {
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

	_, err := w.Write(statusLine)
	return err
}
