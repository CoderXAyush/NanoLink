package base62

import (
	"errors"
	"strings"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base     = uint64(len(alphabet))
)

// Encode converts a database ID to a Base62 string
func Encode(id uint64) string {
	if id == 0 {
		return "a"
	}

	var sb strings.Builder
	for id > 0 {
		remainder := id % base
		sb.WriteByte(alphabet[remainder])
		id /= base
	}

	// Reverse the string because we built it backwards
	chars := []byte(sb.String())
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars)
}

// Decode converts a Base62 string back to a database ID
func Decode(token string) (uint64, error) {
	var id uint64

	for i := 0; i < len(token); i++ {
		pos := strings.IndexByte(alphabet, token[i])
		if pos == -1 {
			return 0, errors.New("invalid character in token")
		}
		id = id*base + uint64(pos)
	}

	return id, nil
}
