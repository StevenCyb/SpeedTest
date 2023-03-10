package status

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalReadinessStatus(t *testing.T) {
	t.Parallel()

	require.False(t, GlobalReadinessStatus())

	ReadinessStatus = true

	require.True(t, GlobalReadinessStatus())
}
