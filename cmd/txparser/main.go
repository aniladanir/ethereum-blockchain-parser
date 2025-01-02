package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aniladanir/ethereum-blockchain-parser/internal/adapters/blockchain"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/adapters/handlers/httphandler"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/adapters/repositories"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/config"
	"github.com/aniladanir/ethereum-blockchain-parser/internal/core/services"
)

var configFile = flag.String("cfg", "./config.json", "provide configuration file")

func main() {
	// parse flags
	flag.Parse()

	// get configs
	cfg, err := config.NewJsonConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	// initialize logger
	logger, loggerCloser, err := InitializeLogger(cfg.GetLogFile(), cfg.GetLogLevel())
	if err != nil {
		log.Fatal(err)
	}
	defer loggerCloser()

	// create parent context
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// listen os exit signals and cancel the parent context if one is received.
	go ListenOsTerminate(parentCancel)

	// create blockchain rpc client
	ethClient := blockchain.NewEthereumClient(time.Second * 5)

	// create repositories
	inMemRepo := repositories.NewInmemTransactionRepository()

	// create services
	txParser := services.NewTransactionParser(inMemRepo, ethClient, logger)

	// create handlers
	serverAddress := fmt.Sprintf("%s:%d", cfg.GetHttpServerIP(), cfg.GetHttpServerPort())
	httpHandler := httphandler.NewHttpHandler(
		serverAddress,
		txParser,
		logger,
	)

	// start processing blockchain
	go func() {
		if err := txParser.ProcessNewBlocks(parentCtx, time.Duration(cfg.GetChainProcessInterval())*time.Millisecond); err != nil {
			logger.Error("process new blocks failed", slog.Any("error", err))
		}
		parentCancel()
	}()

	// start http handler
	go func() {
		if err := httpHandler.Listen(); err != nil {
			logger.Error("error listening http", slog.Any("error", err), slog.String("address", serverAddress))
		}
		parentCancel()
	}()

	// graceful shutdown after context is canceled
	<-parentCtx.Done()
	logger.Info("shutting down txparser gracefully...")

	if err := httpHandler.Shutdown(context.Background()); err != nil {
		logger.Warn("error while shutting down http handler", slog.Any("error", err))
	}
}

func ListenOsTerminate(onSignal func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	onSignal()
}

func InitializeLogger(logFile, level string) (logger *slog.Logger, fileCloser func() error, err error) {
	var sLogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
	case "info":
	case "warn":
	case "error":
	default:
		sLogLevel = slog.LevelDebug
	}

	// create file handler with rotate capabilities
	logFileHandle, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening log file: %w", err)
	}
	logFileCloser := func(fileHandler io.Closer) func() error {
		return func() error {
			return logFileHandle.Close()
		}
	}(logFileHandle)

	// create logger
	if sLogLevel == slog.LevelDebug {
		logger = slog.New(slog.NewTextHandler(logFileHandle, &slog.HandlerOptions{
			Level: sLogLevel,
		}))
	} else {
		logger = slog.New(slog.NewJSONHandler(logFileHandle, &slog.HandlerOptions{
			Level: sLogLevel,
		}))
	}
	slog.SetDefault(logger)

	return logger, logFileCloser, nil
}
