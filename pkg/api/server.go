package api

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Server for status api.
type Server struct {
	logger     *zap.SugaredLogger
	router     *chi.Mux
	httpServer *http.Server
}

// Close the server.
func (s *Server) Close() error {
	if err := s.httpServer.Close(); err != nil {
		return fmt.Errorf("failed to close server: %w", err)
	}

	return nil
}

// ListenAndServe run server (blocking).
func (s *Server) ListenAndServe(readTimeoutSeconds uint, listen string) error {
	httpServer := http.Server{
		Addr:        listen,
		ReadTimeout: time.Second * time.Duration(readTimeoutSeconds),
		Handler:     s.router,
	}
	s.httpServer = &httpServer

	s.logger.Infof("API listening on %s", listen)

	err := httpServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// NewServer create a new server instance.
func NewServer(logger *zap.SugaredLogger, staticFileSystem fs.FS) *Server {
	router := chi.NewRouter()
	server := Server{
		logger: logger,
		router: router,
	}

	router.Handle("/*", http.FileServer(http.FS(staticFileSystem)))

	router.HandleFunc("/latency", server.latencyHandler)

	router.HandleFunc("/downstream", server.downstreamHandler)

	router.HandleFunc("/upstream", server.upstreamHandler)

	return &server
}

func (s *Server) latencyHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "1")
}

func (s *Server) downstreamHandler(resp http.ResponseWriter, req *http.Request) {
	querySize := req.URL.Query().Get("size")
	regex := regexp.MustCompile("^([1-9][0-9]{0,2}|1000)(b|kb|mb)$")

	if querySize == "" || !regex.MatchString(querySize) {
		http.Error(resp, "missing size parameter", http.StatusBadRequest)

		return
	}

	size, err := parseSize(querySize)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)

		return
	}

	resp.Header().Set("Content-Type", "application/octet-stream")
	resp.Header().Set("Content-Length", fmt.Sprintf("%d", size))

	if _, err = io.CopyN(resp, rand.New(rand.NewSource(time.Now().UnixNano())), size); err != nil { //nolint:gosec
		s.logger.Panicf("downstream failed: %w", err)
	}
}

func (s *Server) upstreamHandler(resp http.ResponseWriter, req *http.Request) {
	_, err := io.Copy(io.Discard, req.Body)
	if err != nil {
		s.logger.Error(err)

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			http.Error(resp, "request timed out", http.StatusRequestTimeout)
		} else {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	resp.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(resp, "1")
}
