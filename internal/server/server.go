package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/arunaShakthi/httpfromtcp/internal/request"
	"github.com/arunaShakthi/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("starting TCP listener on %s: %w", addr, err)
	}

	s := &Server{
		listener: listener,
		handler:  handler,
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

	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		body := []byte(fmt.Sprintf("Bad Request: %v\n", err))
		_ = w.WriteStatusLine(response.StatusBadRequest)
		h := response.GetDefaultHeaders(len(body))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody(body)
		return
	}

	if s.handler != nil {
		s.handler(w, req)
	}
}
