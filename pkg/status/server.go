package status

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ProbeFunc defines the format of a condition function.
type ProbeFunc func() bool

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

	s.logger.Infof("Status API listening on %s", listen)

	err := httpServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// NewServer create a new server instance.
/*
 *`readinessConditionFunc`: readiness probes determine whether or not an service ready to serve requests.
 *`livenessConditionFunc`: 	liveness probes determine whether or not an service running healthy.
 */
func NewServer(
	logger *zap.SugaredLogger,
	readinessConditionFuncs, livenessConditionFuncs []ProbeFunc,
) *Server {
	router := chi.NewRouter()
	livenessHandler := func(resp http.ResponseWriter, req *http.Request) {
		for _, livenessConditionFunc := range livenessConditionFuncs {
			if !(livenessConditionFunc)() {
				resp.WriteHeader(http.StatusInternalServerError)
			}
		}

		resp.WriteHeader(http.StatusOK)
	}
	readinessHandler := func(resp http.ResponseWriter, req *http.Request) {
		for _, readinessConditionFunc := range readinessConditionFuncs {
			if !(readinessConditionFunc)() {
				resp.WriteHeader(http.StatusInternalServerError)
			}
		}

		resp.WriteHeader(http.StatusOK)
	}

	router.Get("/ready", readinessHandler)
	router.Get("/health", livenessHandler)

	return &Server{
		logger: logger,
		router: router,
	}
}
