package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading request: %w", err)
	}

	data := string(buf)
	firstLine, _, _ := strings.Cut(data, "\r\n")
	if firstLine == "" {
		firstLine, _, _ = strings.Cut(data, "\n")
	}

	reqLine, err := parseRequestLine(firstLine)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *reqLine,
	}, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, errors.New("invalid request line format: must have 3 parts")
	}

	method := parts[0]
	target := parts[1]
	versionRaw := parts[2]

	if method == "" {
		return nil, errors.New("empty HTTP method")
	}
	for _, ch := range method {
		if !unicode.IsUpper(ch) || ch < 'A' || ch > 'Z' {
			return nil, fmt.Errorf("invalid HTTP method %q: must contain only uppercase ASCII characters", method)
		}
	}

	if !strings.HasPrefix(versionRaw, "HTTP/") {
		return nil, fmt.Errorf("invalid HTTP version format %q", versionRaw)
	}

	version := strings.TrimPrefix(versionRaw, "HTTP/")
	if version != "1.1" {
		return nil, fmt.Errorf("unsupported HTTP version %q: only HTTP/1.1 is supported", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, nil
}
