package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"speedtest/pkg/api"
	"speedtest/pkg/status"
	"speedtest/pkg/utils"
)

//go:embed static/*
var staticFiles embed.FS

//nolint:gochecknoglobals
var (
	logLevel           = utils.GetEnv[string]("LOG_LEVEL", "info", false)
	apiPort            = utils.GetEnv[uint]("API_PORT", "8080", false)
	statusPort         = utils.GetEnv[uint]("STATUS_PORT", "8081", false)
	readTimeoutSeconds = utils.GetEnv[uint]("HTTP_READ_TIMEOUT_SECONDS", "30", false)
)

func main() {
	// setup logger
	logger, err := utils.InitLogger(zap.NewAtomicLevelAt(utils.LevelFromString(logLevel)))
	if err != nil {
		log.Panic("initializing logger failed: ", err)
	}

	logger.Debug("Debug is enabled.")
	logger.Info("Staring service...")

	staticFileSystem, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	// setup status server
	statusServer := status.NewServer(
		logger.Sugar().Named("status"),
		[]status.ProbeFunc{status.GlobalReadinessStatus},
		[]status.ProbeFunc{status.GlobalReadinessStatus},
	)

	go func() {
		if err = statusServer.ListenAndServe(readTimeoutSeconds, fmt.Sprintf(":%d", statusPort)); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Panic(fmt.Sprintf("failed to start status server: %s", err.Error()))
		}
	}()

	defer statusServer.Close()

	// setup api server
	apiServer := api.NewServer(
		logger.Sugar().Named("api"),
		staticFileSystem,
	)

	go func() {
		if err = apiServer.ListenAndServe(readTimeoutSeconds, fmt.Sprintf(":%d", apiPort)); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Panic(fmt.Sprintf("failed to start status server: %s", err.Error()))
		}
	}()

	defer apiServer.Close()

	// set flag for readiness
	status.ReadinessStatus = true

	logger.Info("Service is running")

	// listen shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
