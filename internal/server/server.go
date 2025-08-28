package server

import (
	"fmt"
	"io"
	"net"

	"github.com/devasherr/tcp-http/internal/response"
)

type Server struct {
	closed bool
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	response.WriteStatusLine(conn, response.StatusOk)
	response.WriteHeaders(conn, headers)
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

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("error: ", err)
		return nil, err
	}

	s := &Server{
		closed: false,
	}

	go runServer(s, listener)
	return s, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
