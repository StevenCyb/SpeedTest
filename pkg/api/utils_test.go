package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSizeParsing(t *testing.T) {
	t.Parallel()

	size, err := parseSize("5b")
	require.NoError(t, err)
	require.Equal(t, int64(5), size)

	size, err = parseSize("24kb")
	require.NoError(t, err)
	require.Equal(t, int64(24576), size)

	size, err = parseSize("100mb")
	require.NoError(t, err)
	require.Equal(t, int64(104857600), size)
}
