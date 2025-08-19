package request

import (
	"errors"
	"io"
	"strings"

	"fmt"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var ERROR_INVALID_REQUETS_LINE = fmt.Errorf("invalid request line")
var SEPARTOR = "\r\n"

func parseRequestLine(data string) (*RequestLine, string, error) {
	idx := strings.Index(data, SEPARTOR)
	firstPart := data[:idx]
	restOfMsg := data[idx+len(SEPARTOR):]

	messageComponents := strings.Split(firstPart, " ")
	if len(messageComponents) != 3 {
		return nil, restOfMsg, ERROR_INVALID_REQUETS_LINE
	}

	httpVersion := strings.Split(messageComponents[2], "/")
	if len(httpVersion) != 2 || httpVersion[0] != "HTTP" || httpVersion[1] != "1.1" {
		return nil, restOfMsg, ERROR_INVALID_REQUETS_LINE
	}

	return &RequestLine{
		Method:        messageComponents[0],
		RequestTarget: messageComponents[1],
		HttpVersion:   httpVersion[1],
	}, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("unable to io.ReadAll"), err)
	}

	str := string(data)
	reqLine, _, err := parseRequestLine(str)
	if err != nil {
		return nil, err
	}

	return &Request{RequestLine: *reqLine}, nil
}
