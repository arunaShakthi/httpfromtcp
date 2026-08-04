package response

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteStatusLine(t *testing.T) {
	t.Run("Status 200 OK", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteStatusLine(&buf, StatusOK)
		require.NoError(t, err)
		assert.Equal(t, "HTTP/1.1 200 OK\r\n", buf.String())
	})

	t.Run("Status 400 Bad Request", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteStatusLine(&buf, StatusBadRequest)
		require.NoError(t, err)
		assert.Equal(t, "HTTP/1.1 400 Bad Request\r\n", buf.String())
	})

	t.Run("Status 500 Internal Server Error", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteStatusLine(&buf, StatusInternalServerError)
		require.NoError(t, err)
		assert.Equal(t, "HTTP/1.1 500 Internal Server Error\r\n", buf.String())
	})

	t.Run("Unknown status code", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteStatusLine(&buf, StatusCode(418))
		require.NoError(t, err)
		assert.Equal(t, "HTTP/1.1 418 \r\n", buf.String())
	})
}

func TestGetDefaultHeaders(t *testing.T) {
	h := GetDefaultHeaders(13)
	assert.Equal(t, "13", h.Get("Content-Length"))
	assert.Equal(t, "close", h.Get("Connection"))
	assert.Equal(t, "text/plain", h.Get("Content-Type"))
}

func TestResponseWriter(t *testing.T) {
	t.Run("Valid order", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf)

		err := w.WriteStatusLine(StatusOK)
		require.NoError(t, err)

		h := GetDefaultHeaders(12)
		err = w.WriteHeaders(h)
		require.NoError(t, err)

		n, err := w.WriteBody([]byte("hello world\n"))
		require.NoError(t, err)
		assert.Equal(t, 12, n)
	})

	t.Run("Out of order WriteHeaders before status line", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf)

		h := GetDefaultHeaders(10)
		err := w.WriteHeaders(h)
		require.Error(t, err)
	})

	t.Run("Out of order WriteBody before headers", func(t *testing.T) {
		var buf bytes.Buffer
		w := NewWriter(&buf)

		_ = w.WriteStatusLine(StatusOK)
		_, err := w.WriteBody([]byte("data"))
		require.Error(t, err)
	})
}
