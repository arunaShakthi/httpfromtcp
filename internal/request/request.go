package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const bufferSize = 8

type parserState int

const (
	stateInitialized parserState = iota
	stateDone
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		reqLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *reqLine
		r.state = stateDone
		return n, nil

	case stateDone:
		return 0, errors.New("error: trying to read data in a done state")

	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	str := string(data)
	idx := strings.Index(str, "\r\n")
	if idx == -1 {
		return nil, 0, nil
	}

	line := str[:idx]
	bytesConsumed := idx + 2

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, 0, errors.New("invalid request line format: must have 3 parts")
	}

	method := parts[0]
	target := parts[1]
	versionRaw := parts[2]

	if method == "" {
		return nil, 0, errors.New("empty HTTP method")
	}
	for _, ch := range method {
		if !unicode.IsUpper(ch) || ch < 'A' || ch > 'Z' {
			return nil, 0, fmt.Errorf("invalid HTTP method %q: must contain only uppercase ASCII characters", method)
		}
	}

	if !strings.HasPrefix(versionRaw, "HTTP/") {
		return nil, 0, fmt.Errorf("invalid HTTP version format %q", versionRaw)
	}

	version := strings.TrimPrefix(versionRaw, "HTTP/")
	if version != "1.1" {
		return nil, 0, fmt.Errorf("unsupported HTTP version %q: only HTTP/1.1 is supported", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, bytesConsumed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{state: stateInitialized}

	for req.state != stateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if n > 0 {
			readToIndex += n
		}

		if readToIndex > 0 {
			nParsed, parseErr := req.parse(buf[:readToIndex])
			if parseErr != nil {
				return nil, parseErr
			}
			if nParsed > 0 {
				copy(buf, buf[nParsed:readToIndex])
				readToIndex -= nParsed
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != stateDone {
					req.state = stateDone
				}
				break
			}
			return nil, err
		}
	}

	return req, nil
}
