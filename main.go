package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"everyday_keywords/internal/cleaner"
	"everyday_keywords/internal/collector"
	"everyday_keywords/internal/db"
	"everyday_keywords/internal/handler"
	"everyday_keywords/internal/logger"
	"everyday_keywords/internal/notifier"
)

type Config struct {
	Port        int    `json:"port"`
	DBPath      string `json:"db_path"`
	LogFile     string `json:"log_file"`
	WebhookURL  string `json:"webhook_url"`
	WebhookHMAC string `json:"webhook_hmac"`
}

func defaultConfig() Config {
	return Config{
		Port:        8080,
		DBPath:      "/data/keywords.db",
		LogFile:     "/data/requests.log",
		WebhookURL:  "",
		WebhookHMAC: "",
	}
}

func loadConfig() Config {
	cfg := defaultConfig()
	configPath := "/app/config/config.json"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Printf("[config] cannot create config dir: %v, using defaults", err)
			return cfg
		}
		data, _ := json.MarshalIndent(cfg, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			log.Printf("[config] cannot write default config: %v, using defaults", err)
		} else {
			log.Printf("[config] created default config at %s", configPath)
		}
		return cfg
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("[config] cannot read config: %v, using defaults", err)
		return cfg
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		log.Printf("[config] invalid config JSON: %v, using defaults", err)
		return cfg
	}

	if fileCfg.Port != 0 {
		cfg.Port = fileCfg.Port
	}
	if fileCfg.DBPath != "" {
		cfg.DBPath = fileCfg.DBPath
	}
	if fileCfg.LogFile != "" {
		cfg.LogFile = fileCfg.LogFile
	}
	if fileCfg.WebhookURL != "" {
		cfg.WebhookURL = fileCfg.WebhookURL
	}
	if fileCfg.WebhookHMAC != "" {
		cfg.WebhookHMAC = fileCfg.WebhookHMAC
	}

	return cfg
}

func main() {
	cfg := loadConfig()

	log.Printf("[config] port=%d, db=%s, log=%s, webhook=%s, hmac=%v",
		cfg.Port, cfg.DBPath, cfg.LogFile, cfg.WebhookURL, cfg.WebhookHMAC != "")

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("[db] open failed: %v", err)
	}
	defer database.Close()
	log.Printf("[db] connected to %s", cfg.DBPath)

	logger := logger.New(cfg.LogFile)
	defer logger.Close()

	notifier := notifier.New(cfg.WebhookURL, cfg.WebhookHMAC)

	col := collector.New(database)

	total, _ := database.Count()
	if total < 100 {
		log.Printf("[init] keyword count (%d) < 100, collecting initial pool...", total)
		_, errs := col.CollectUntil(100)
		if len(errs) > 0 {
			for _, e := range errs {
				log.Printf("[init] collect error: %v", e)
			}
		}
		newTotal, _ := database.Count()
		log.Printf("[init] keyword pool now has %d keywords", newTotal)
	}

	cleaner := cleaner.New(database)
	cleaner.Start(30 * time.Minute)
	log.Printf("[cleaner] started, interval=30m")

	h := handler.New(database, col, notifier, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/keywords", h.ServeHTTP)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[server] listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[server] %v", err)
	}
}