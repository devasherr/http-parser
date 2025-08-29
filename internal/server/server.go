package server

import (
	"fmt"
	"io"
	"net"

	"github.com/devasherr/tcp-http/internal/request"
	"github.com/devasherr/tcp-http/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed  bool
	handler Handler
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, req)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed {
			return
		}

		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		go runConnection(s, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("error: ", err)
		return nil, err
	}

	s := &Server{
		closed:  false,
		handler: handler,
	}

	go runServer(s, listener)
	return s, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
