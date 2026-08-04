package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/arunaShakthi/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("starting TCP listener on %s: %w", addr, err)
	}

	s := &Server{
		listener: listener,
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Accept error: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error writing status line: %v\n", err)
		return
	}

	headers := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error writing headers: %v\n", err)
		return
	}
}
