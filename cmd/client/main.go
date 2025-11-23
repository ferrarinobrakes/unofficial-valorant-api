package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/config"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/lcu"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/logging"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/protocol"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/valorant"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/version"
)

func main() {
	logger, err := logging.NewLogger()
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting Valorant API Client Node", "version", version.Version)

	cfg, err := config.LoadClientConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("configuration loaded", "clientID", cfg.ClientID, "masterAddress", cfg.MasterAddress)

	var lockfile *lcu.LockfileData

	for {
		lockfile, err = lcu.ReadLockfile()
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	logger.Info("riot client detected", "port", lockfile.Port)

	lcuClient := lcu.NewClient(lockfile, logger)
	valClient := valorant.NewClient(logger)
	resolver := lcu.NewResolver(lcuClient, valClient, logger)
	client := protocol.NewClient(cfg.MasterAddress, cfg.ClientID, version.Version, resolver, logger)

	for {
		err = client.Connect()
		if err == nil {
			break
		}
		logger.Errorw("failed to connect to master, retrying...", "error", err)
		time.Sleep(5 * time.Second)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down...")
		client.Stop()
		os.Exit(0)
	}()

	if err := client.Run(); err != nil {
		logger.Errorw("client error", "error", err)
		os.Exit(1)
	}

	logger.Info("client stopped")
}
