package status

import (
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"speedtest/pkg/utils"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

//nolint:noctx
func TestServer(t *testing.T) {
	t.Parallel()

	if _, isPipeline := os.LookupEnv("CI"); isPipeline {
		t.Skip("Running server is not possible in CI")

		return
	}

	logger, err := utils.InitLogger(zap.NewAtomicLevelAt(zap.WarnLevel))
	require.NoError(t, err)

	status := false
	readinessConditionFunc := ProbeFunc(func() bool { return status })
	livenessConditionFunc := ProbeFunc(func() bool { return status })
	server := NewServer(
		logger.Sugar(),
		[]ProbeFunc{readinessConditionFunc},
		[]ProbeFunc{livenessConditionFunc},
	)

	go func(server *Server) {
		require.True(t, errors.Is(server.ListenAndServe(5, ":8886"), http.ErrServerClosed))
	}(server)

	res, err := http.Get("http://localhost:8886/ready")
	require.NoError(t, err)
	require.Equal(t, 500, res.StatusCode)
	res.Body.Close()

	res, err = http.Get("http://localhost:8886/health")
	require.NoError(t, err)
	require.Equal(t, 500, res.StatusCode)
	res.Body.Close()

	status = true
	res, err = http.Get("http://localhost:8886/ready")
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	res.Body.Close()

	res, err = http.Get("http://localhost:8886/health")
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	res.Body.Close()

	time.Sleep(3 * time.Second)
	server.Close()
}
