package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arunaShakthi/httpfromtcp/internal/request"
	"github.com/arunaShakthi/httpfromtcp/internal/response"
	"github.com/arunaShakthi/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		var statusCode response.StatusCode
		var body string

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			statusCode = response.StatusBadRequest
			body = "<html><body><h1>400 Bad Request</h1><p>Your problem is not my problem</p></body></html>\n"
		case "/myproblem":
			statusCode = response.StatusInternalServerError
			body = "<html><body><h1>500 Internal Server Error</h1><p>Woopsie, my bad</p></body></html>\n"
		default:
			statusCode = response.StatusOK
			body = "<html><body><h1>200 OK</h1><p>All good, frfr</p></body></html>\n"
		}

		bodyBytes := []byte(body)
		_ = w.WriteStatusLine(statusCode)

		h := response.GetDefaultHeaders(len(bodyBytes))
		h["content-type"] = "text/html"

		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody(bodyBytes)
	}

	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
