package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/arunaShakthi/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	writerStateStart writerState = iota
	writerStateStatusLineWritten
	writerStateHeadersWritten
	writerStateBodyWritten
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateStart,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStart {
		return errors.New("error: WriteStatusLine called out of order")
	}
	err := WriteStatusLine(w.writer, statusCode)
	if err == nil {
		w.state = writerStateStatusLineWritten
	}
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != writerStateStatusLineWritten {
		return errors.New("error: WriteHeaders called out of order")
	}
	err := WriteHeaders(w.writer, h)
	if err == nil {
		w.state = writerStateHeadersWritten
	}
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateHeadersWritten && w.state != writerStateBodyWritten {
		return 0, errors.New("error: WriteBody called out of order")
	}
	w.state = writerStateBodyWritten
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateHeadersWritten && w.state != writerStateBodyWritten {
		return 0, errors.New("error: WriteChunkedBody called out of order")
	}
	if len(p) == 0 {
		return 0, nil
	}
	w.state = writerStateBodyWritten

	chunkHeader := fmt.Sprintf("%X\r\n", len(p))
	if _, err := w.writer.Write([]byte(chunkHeader)); err != nil {
		return 0, err
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}
	if _, err := w.writer.Write([]byte("\r\n")); err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateHeadersWritten && w.state != writerStateBodyWritten {
		return 0, errors.New("error: WriteChunkedBodyDone called out of order")
	}
	w.state = writerStateBodyWritten
	return w.writer.Write([]byte("0\r\n\r\n"))
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}

	var statusLine string
	if reason != "" {
		statusLine = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)
	} else {
		statusLine = fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)
	}

	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for k, v := range h {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}
