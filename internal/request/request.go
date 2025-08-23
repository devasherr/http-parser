package request

import (
	"bytes"
	"io"

	"fmt"

	"github.com/devasherr/tcp-http/internal/headers"
)

var ERROR_INVALID_REQUETS_LINE = fmt.Errorf("invalid request line")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("request in error state")
var SEPARTOR = []byte("\r\n")

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	State       parserState
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, SEPARTOR)
	if idx == -1 {
		return nil, 0, nil
	}

	firstPart := data[:idx]
	read := idx + len(SEPARTOR)

	httpParts := bytes.Split(firstPart, []byte(" "))
	if len(httpParts) != 3 {
		return nil, 0, ERROR_INVALID_REQUETS_LINE
	}

	httpVersion := bytes.Split(httpParts[2], []byte("/"))
	if len(httpVersion) != 2 || string(httpVersion[0]) != "HTTP" || string(httpVersion[1]) != "1.1" {
		return nil, 0, ERROR_INVALID_REQUETS_LINE
	}

	return &RequestLine{
		Method:        string(httpParts[0]),
		RequestTarget: string(httpParts[1]),
		HttpVersion:   string(httpVersion[1]),
	}, read, nil
}

func (r *Request) done() bool {
	return r.State == StateDone
}

func (r *Request) error() bool {
	return r.State == StateError
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		switch r.State {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.State = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				r.State = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("i don't know how we got here")
		}
	}

	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
