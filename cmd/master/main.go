package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ferrarinobrakes/unofficial-valorant-api/gen/v1connect"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/api"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/cache"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/config"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/db"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/logging"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/protocol"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/version"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	logger, err := logging.NewLogger()
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting Valorant API Master Server")

	cfg := config.LoadMasterConfig()
	logger.Info("configuration loaded", zap.Int("tcpPort", cfg.TCPPort), zap.Int("apiPort", cfg.APIPort))

	database, err := db.New(cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to initialize database", zap.Error(err))
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("database initialized", zap.String("path", cfg.DatabasePath))

	memCache := cache.New(cfg.CacheTTL)
	logger.Info("cache initialized", zap.Duration("ttl", cfg.CacheTTL))

	tcpServer, err := protocol.NewServer(cfg.TCPPort, logger)
	if err != nil {
		logger.Error("failed to start TCP server", zap.Error(err))
		os.Exit(1)
	}
	defer tcpServer.Stop()

	go func() {
		if err := tcpServer.Start(); err != nil {
			logger.Error("tcp server error", zap.Error(err))
		}
	}()

	apiService := api.NewService(tcpServer, memCache, database, logger)

	mux := http.NewServeMux()

	path, handler := v1connect.NewValorantAPIHandler(apiService)
	mux.Handle(path, handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK - %d clients connected", tcpServer.GetClientCount())
	})

	addr := fmt.Sprintf(":%d", cfg.APIPort)
	logger.Info("starting API server", zap.Int("port", cfg.APIPort), zap.String("version", version.Version))

	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down...")
		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("api server error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
