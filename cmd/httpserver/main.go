package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/arunaShakthi/httpfromtcp/internal/request"
	"github.com/arunaShakthi/httpfromtcp/internal/response"
	"github.com/arunaShakthi/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		target := req.RequestLine.RequestTarget

		if strings.HasPrefix(target, "/httpbin/") {
			subPath := strings.TrimPrefix(target, "/httpbin/")
			targetURL := "https://httpbingo.org/" + subPath

			resp, err := http.Get(targetURL)
			if err != nil {
				log.Printf("proxy GET error: %v\n", err)
				_ = w.WriteStatusLine(response.StatusInternalServerError)
				h := response.GetDefaultHeaders(0)
				_ = w.WriteHeaders(h)
				return
			}
			defer resp.Body.Close()

			_ = w.WriteStatusLine(response.StatusOK)

			h := response.GetDefaultHeaders(0)
			delete(h, "content-length")
			h["transfer-encoding"] = "chunked"
			h["connection"] = "close"
			if ct := resp.Header.Get("Content-Type"); ct != "" {
				h["content-type"] = ct
			}

			_ = w.WriteHeaders(h)

			buf := make([]byte, 1024)
			for {
				n, rErr := resp.Body.Read(buf)
				if n > 0 {
					log.Printf("proxy read %d bytes\n", n)
					_, _ = w.WriteChunkedBody(buf[:n])
				}
				if rErr != nil {
					if errors.Is(rErr, io.EOF) {
						break
					}
					log.Printf("error reading proxy body: %v\n", rErr)
					break
				}
			}
			_, _ = w.WriteChunkedBodyDone()
			return
		}

		var statusCode response.StatusCode
		var body string

		switch target {
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
