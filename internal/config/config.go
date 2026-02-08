package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DataDir     string          `json:"data_dir"`
	CacheDir    string          `json:"cache_dir"`
	AIWorker    AIWorkerConfig  `json:"ai_worker"`
	Sentinel    SentinelConfig  `json:"sentinel"`
	Landsat     LandsatConfig   `json:"landsat"`
	Planet      PlanetConfig    `json:"planet"`
	Monitoring  MonitorConfig   `json:"monitoring"`
}

type AIWorkerConfig struct {
	Address string `json:"address"`
	UseTLS  bool   `json:"use_tls"`
}

type SentinelConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type LandsatConfig struct {
	USGSUsername string `json:"usgs_username"`
	USGSPassword string `json:"usgs_password"`
}

type PlanetConfig struct {
	APIKey string `json:"api_key"`
}

type MonitorConfig struct {
	CheckInterval string `json:"check_interval"`
	AlertWebhook  string `json:"alert_webhook"`
}

func Load() *Config {
	cfg := &Config{
		DataDir:  "data",
		CacheDir: filepath.Join(os.TempDir(), "orbital-eye"),
		AIWorker: AIWorkerConfig{Address: "localhost:50051"},
	}

	// Try loading from config file
	home, _ := os.UserHomeDir()
	paths := []string{
		"orbital-eye.json",
		filepath.Join(home, ".config", "orbital-eye", "config.json"),
	}
	for _, p := range paths {
		if f, err := os.Open(p); err == nil {
			json.NewDecoder(f).Decode(cfg)
			f.Close()
			break
		}
	}

	return cfg
}
