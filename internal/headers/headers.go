package headers

import (
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	str := string(data)
	idx := strings.Index(str, "\r\n")
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	line := str[:idx]
	bytesConsumed := idx + 2

	keyRaw, valRaw, found := strings.Cut(line, ":")
	if !found {
		return 0, false, errors.New("invalid header line: missing colon")
	}

	if len(keyRaw) == 0 {
		return 0, false, errors.New("invalid header line: empty field name")
	}

	if strings.TrimSpace(keyRaw) != keyRaw || strings.ContainsAny(keyRaw, " \t") {
		return 0, false, errors.New("invalid header line: spacing in field name")
	}

	for _, ch := range keyRaw {
		if !isValidHeaderNameChar(ch) {
			return 0, false, errors.New("invalid character in header field name")
		}
	}

	val := strings.TrimSpace(valRaw)
	key := strings.ToLower(keyRaw)

	if existing, ok := h[key]; ok {
		h[key] = existing + ", " + val
	} else {
		h[key] = val
	}

	return bytesConsumed, false, nil
}

func isValidHeaderNameChar(ch rune) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
		return true
	}
	switch ch {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}
