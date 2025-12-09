package auth

import (
	"errors"
	"strconv"
)

func cloneSlice[T any](src []T) []T {
	if len(src) == 0 {
		return nil
	}
	out := make([]T, len(src))
	copy(out, src)
	return out
}

func parseUint32(v string) (uint32, error) {
	if v == "" {
		return 0, errors.New("empty")
	}
	u, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(u), nil
}

// Errors placeholder
var (
	ErrInvalidAction = errors.New("invalid action")
)
