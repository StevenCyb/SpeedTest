package api

import (
	"errors"
	"strconv"
	"strings"
)

//nolint:varnamelen
const (
	b  = 1
	kb = 1024 * b
	mb = 1024 * kb
)

var ErrInvalidSizeParameter = errors.New("invalid size parameter")

func parseSize(sizeStr string) (int64, error) {
	var (
		multiplier int64
		err        error
	)

	switch {
	case strings.HasSuffix(sizeStr, "kb"):
		multiplier = kb
		sizeStr = strings.TrimSuffix(sizeStr, "kb")
	case strings.HasSuffix(sizeStr, "mb"):
		multiplier = mb
		sizeStr = strings.TrimSuffix(sizeStr, "mb")
	case strings.HasSuffix(sizeStr, "b"):
		multiplier = b
		sizeStr = strings.TrimSuffix(sizeStr, "b")
	default:
		multiplier = 1
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidSizeParameter
	}

	return size * multiplier, nil
}
