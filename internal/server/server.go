package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/arunaShakthi/httpfromtcp/internal/request"
	"github.com/arunaShakthi/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    fmt.Sprintf("Bad Request: %v\n", err),
		}
		WriteHandlerError(conn, hErr)
		return
	}

	var bodyBuf bytes.Buffer
	if s.handler != nil {
		hErr := s.handler(&bodyBuf, req)
		if hErr != nil {
			WriteHandlerError(conn, hErr)
			return
		}
	}

	bodyBytes := bodyBuf.Bytes()
	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error writing status line: %v\n", err)
		return
	}

	headers := response.GetDefaultHeaders(len(bodyBytes))
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error writing headers: %v\n", err)
		return
	}

	if len(bodyBytes) > 0 {
		if _, err := conn.Write(bodyBytes); err != nil {
			log.Printf("Error writing response body: %v\n", err)
		}
	}
}

func WriteHandlerError(w io.Writer, hErr *HandlerError) {
	if hErr == nil {
		return
	}
	body := []byte(hErr.Message)
	_ = response.WriteStatusLine(w, hErr.StatusCode)
	headers := response.GetDefaultHeaders(len(body))
	_ = response.WriteHeaders(w, headers)
	if len(body) > 0 {
		_, _ = w.Write(body)
	}
}
