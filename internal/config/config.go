package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type MasterConfig struct {
	TCPPort      int
	APIPort      int
	DatabasePath string
	CacheTTL     time.Duration
}

func LoadMasterConfig() *MasterConfig {
	return &MasterConfig{
		TCPPort:      getEnvInt("TCP_PORT", 8080),
		APIPort:      getEnvInt("API_PORT", 8081),
		DatabasePath: getEnv("DATABASE_PATH", "./data/valorant.db"),
		CacheTTL:     time.Duration(getEnvInt("CACHE_TTL_MINUTES", 60)) * time.Minute,
	}
}

type ClientConfig struct {
	MasterAddress string `json:"master_address"`
	ClientID      string `json:"client_id"`
	LogLevel      string `json:"log_level"`
}

func LoadClientConfig() (*ClientConfig, error) {
	data, err := os.ReadFile("config.json")
	if err == nil {
		var cfg ClientConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config.json: %w", err)
		}

		if cfg.ClientID == "" {
			hostname, _ := os.Hostname()
			cfg.ClientID = fmt.Sprintf("client-%s", hostname)
		}

		if cfg.LogLevel == "" {
			cfg.LogLevel = "info"
		}
		os.Setenv("LOG_LEVEL", cfg.LogLevel)

		return &cfg, nil
	}

	cfg := &ClientConfig{
		MasterAddress: getEnv("MASTER_ADDRESS", "localhost:8080"),
		ClientID:      getEnv("CLIENT_ID", ""),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	if cfg.ClientID == "" {
		hostname, _ := os.Hostname()
		cfg.ClientID = fmt.Sprintf("client-%s", hostname)
	}

	os.Setenv("LOG_LEVEL", cfg.LogLevel)

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
