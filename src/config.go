package main

import (
	"os"
	"strconv"
)

type Config struct {
	SteamAPIKey   string
	ProxyUser     string
	ProxyPassword string
	ProxyHost     string
	ProxyPort     string
	LockTimeout   int64
	MaxWorkers    int
	DatabasePath  string
	SeedSteamID   int64
}

func LoadConfig() Config {
	lockTimeout, _ := strconv.ParseInt(getEnv("LOCK_TIMEOUT", "300"), 10, 64)
	maxWorkers, _ := strconv.Atoi(getEnv("MAX_WORKERS", "5"))
	seedSteamID, _ := strconv.ParseInt(os.Getenv("SEED_STEAM_ID"), 10, 64)

	return Config{
		SteamAPIKey:   os.Getenv("STEAM_API_KEY"),
		ProxyUser:     os.Getenv("PROXY_USER"),
		ProxyPassword: os.Getenv("PROXY_PASSWORD"),
		ProxyHost:     os.Getenv("PROXY_HOST"),
		ProxyPort:     os.Getenv("PROXY_PORT"),
		LockTimeout:   lockTimeout,
		MaxWorkers:    maxWorkers,
		DatabasePath:  getEnv("DATABASE_PATH", "steamer.db"),
		SeedSteamID:   seedSteamID,
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
