# Custom HTTP/1.1 Server (HTTP from TCP in Go)

A custom, low-level HTTP/1.1 web server built from scratch directly on top of Go's `net.TCPListener` (`net.Listen("tcp", ...)`), without using standard library HTTP packages (`net/http`) for request parsing or server handling.

---

## 🚀 What Was Done & Implemented

### 1. Networking Basics (TCP vs UDP)
- **TCP Listener (`cmd/tcplistener`)**: Built a connection-oriented socket listener on port `:42069` using `net.Listen` and `listener.Accept()`. Demonstrated connection handshakes, data streaming, and socket termination.
- **UDP Sender (`cmd/udpsender`)**: Built a connectionless UDP client using `net.ResolveUDPAddr` and `net.DialUDP` to demonstrate packet yeeting without requiring connection state or receiver acknowledgment.

### 2. Streaming HTTP Request Parser State Machine (`internal/request`)
- **State Machine Architecture**: Created an internal state machine (`stateInitialized` → `stateParsingHeaders` → `stateParsingBody` → `stateDone`) to process incoming TCP bytes as they arrive over the wire.
- **Dynamic Chunked Buffering**: Handles partial network reads and small buffer sizes (tested with buffer sizes down to 8 bytes) by shifting unparsed data efficiently.
- **Request Line Parsing**: Extracts and validates the HTTP Method (uppercase ASCII validation), Target Path, and Version (`HTTP/1.1`).
- **Header Parsing Integration**: Processes header lines up to the empty CRLF line (`\r\n\r\n`).
- **Body Parsing**: Reads raw body payloads based on the `Content-Length` header, handling chunked network reads and erroring on unexpected EOF or payload length mismatches.

### 3. Header Data Structure & Validation (`internal/headers`)
- **Headers Map (`Headers map[string]string`)**: Custom map for managing HTTP field lines.
- **Case-Insensitive Access**: Automatically lowercases all keys upon storage and provides a `.Get(key)` helper for case-insensitive header lookups.
- **RFC 9110 Compliance**:
  - Validates `tchar` token characters for header keys.
  - Rejects leading whitespace before field names or spaces before colons.
  - Trims optional whitespace around values.
  - Automatically joins duplicate headers with a comma and space (`", "`).

### 4. HTTP Response Writer & State Machine (`internal/response`)
- **Status Line Writing**: Formats status lines for standard codes (`200 OK`, `400 Bad Request`, `500 Internal Server Error`) and custom codes.
- **Default Headers Generator**: Generates standard response headers (`Content-Length`, `Connection: close`, `Content-Type`).
- **Response `Writer` Struct**: Encapsulates `io.Writer` and enforces strict calling sequence (`WriteStatusLine` → `WriteHeaders` → `WriteBody` / `WriteChunkedBody` / `WriteTrailers`). Returns explicit errors if out-of-order calls occur.
- **Chunked Transfer-Encoding**: Supports streaming responses using hexadecimal chunk sizes (`WriteChunkedBody`) and the terminal zero chunk (`WriteChunkedBodyDone`).
- **HTTP Trailer Headers (`WriteTrailers`)**: Appends post-body metadata trailers (`X-Content-SHA256`, `X-Content-Length`) after the zero-length chunk.

### 5. Concurrent Server & Graceful Shutdown (`internal/server`)
- **Concurrent Handler**: Accepts incoming TCP connections in a loop and handles each connection concurrently in a separate goroutine (`go s.handle(conn)`).
- **Graceful Shutdown**: Tracks server closed state with `sync/atomic.Bool` and catches OS signals (`SIGINT`, `SIGTERM`) in `main.go` to close listeners cleanly.

### 6. Reverse Proxying & Streaming Demonstration (`cmd/httpserver`)
- **Reverse Proxy Route (`/httpbin/*`)**: Proxies requests to `https://httpbingo.org/*`.
- **Real-Time Chunked Streaming**: Streams response chunks from `httpbingo.org` back to the client in real-time.
- **Dynamic Trailer Calculation**: Accumulates full response payloads, computes SHA-256 digests (`crypto/sha256`), and emits `X-Content-SHA256` and `X-Content-Length` trailers.

---

## 📁 Repository Structure

```text
├── cmd/
│   ├── httpserver/       # Main HTTP server binary with routing & proxying
│   ├── tcplistener/      # Low-level TCP socket listener utility
│   └── udpsender/        # Interactive UDP datagram sender
├── internal/
│   ├── headers/          # Case-insensitive HTTP Header map & RFC parser
│   ├── request/          # Streaming HTTP Request parser state machine
│   ├── response/         # Status codes, headers, response Writer & chunked encoding
│   └── server/           # Concurrent TCP HTTP server framework
├── go.mod                # Module definition
└── README.md             # Documentation
```

---

## 🧪 Testing & Verification

### Run Unit Tests
Run unit test suites across all packages:
```bash
go test -v ./...
```

### Run HTTP Server
Start the HTTP server on port `42069`:
```bash
go run ./cmd/httpserver
```

### Test Routes

1. **Standard HTML Response**:
   ```bash
   curl -i http://localhost:42069/
   ```

2. **Error Routes**:
   ```bash
   curl -i http://localhost:42069/yourproblem   # 400 Bad Request
   curl -i http://localhost:42069/myproblem     # 500 Internal Server Error
   ```

3. **Proxy & Raw Chunked Response with Trailers**:
   ```bash
   curl -v --raw http://localhost:42069/httpbin/range/4096
   ```

4. **Raw TCP Socket Streaming (using Netcat)**:
   ```bash
   printf "GET /httpbin/stream/3 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
   ```
