package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"speedtest/pkg/utils"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockFS struct{}

func (m *mockFS) Open(name string) (fs.File, error) {
	switch name {
	case ".":
		return &fakeFile{name: "static", isDir: true, content: []byte("index.html")}, nil
	case "index.html":
		return &fakeFile{name: "index.html", content: []byte("<html><head></head><body></body></html>")}, nil
	}

	return nil, fmt.Errorf("file not found: %s", name) //nolint:goerr113
}

type fakeFile struct {
	name    string
	content []byte
	pos     int
	isDir   bool
}

func (f *fakeFile) Stat() (fs.FileInfo, error) {
	return fileInfo{
		name:    f.name,
		size:    int64(len(f.content)),
		mode:    fs.ModePerm,
		modTime: time.Now(),
		isDir:   f.isDir,
	}, nil
}

func (f *fakeFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.content) {
		return 0, io.EOF
	}

	n := copy(p, f.content[f.pos:])
	f.pos += n

	return n, nil
}

func (f *fakeFile) Close() error {
	return nil
}

type fileInfo struct {
	modTime time.Time
	name    string
	size    int64
	isDir   bool
	mode    fs.FileMode
}

func (fi fileInfo) Name() string {
	return fi.name
}

func (fi fileInfo) Size() int64 {
	return fi.size
}

func (fi fileInfo) Mode() fs.FileMode {
	return fi.mode
}

func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi fileInfo) IsDir() bool {
	return fi.isDir
}

func (fi fileInfo) Sys() interface{} {
	return nil
}

//nolint:noctx
func TestServer(t *testing.T) {
	t.Parallel()

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip()
	}

	if _, isPipeline := os.LookupEnv("CI"); isPipeline {
		t.Skip("Running server is not possible in CI")

		return
	}

	logger, err := utils.InitLogger(zap.NewAtomicLevelAt(zap.WarnLevel))
	require.NoError(t, err)

	server := NewServer(
		logger.Sugar(),
		&mockFS{},
	)

	go func(server *Server) {
		require.True(t, errors.Is(server.ListenAndServe(5, ":8080"), http.ErrServerClosed))
	}(server)

	resp, err := http.Get("http://localhost:8080/")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()

	time.Sleep(3 * time.Second)
	server.Close()
}

func TestLatencyHandler(t *testing.T) {
	t.Parallel()

	server := NewServer(nil, nil)
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, "/latency",
		nil)
	recorder := httptest.NewRecorder()

	require.NoError(t, err)
	server.latencyHandler(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestDownstreamHandler(t *testing.T) {
	t.Parallel()

	server := NewServer(nil, nil)

	t.Run("ValidSize", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodGet, "/downstream?size=8b",
			nil)
		recorder := httptest.NewRecorder()

		require.NoError(t, err)
		server.downstreamHandler(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "8", recorder.Header().Get("Content-Length"))

		body, err := io.ReadAll(recorder.Body)
		require.NoError(t, err)
		require.Equal(t, 8, len(body))
	})

	t.Run("InvalidSize", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(context.Background(),
			http.MethodGet, "/downstream?size=10000mb",
			nil)
		recorder := httptest.NewRecorder()

		require.NoError(t, err)
		server.downstreamHandler(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestUpstreamHandler(t *testing.T) {
	t.Parallel()

	// Create a new request with a multipart form body.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "filename")
	require.NoError(t, err)

	// Write one byte to the file part.
	_, err = part.Write([]byte{0})
	require.NoError(t, err)

	// Close the multipart writer and set the content type header.
	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, "/upstream",
		body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()

	// Call the handler function.
	server := NewServer(nil, nil)
	server.upstreamHandler(recorder, req)

	// Check the response.
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "1", recorder.Body.String())
}
