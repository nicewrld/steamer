// src/main.go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {

	err := profiler.Start(
		profiler.WithService(os.Getenv("DD_SERVICE")),
		profiler.WithEnv(os.Getenv("DD_ENV")),
		profiler.WithVersion(os.Getenv("DD_VERSION")),
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
			// The profiles below are disabled by default to keep overhead
			// low, but can be enabled as needed.
			profiler.BlockProfile,
			profiler.MutexProfile,
			profiler.GoroutineProfile,
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer profiler.Stop()

	config := LoadConfig()

	// Validate essential configurations
	if config.SteamAPIKey == "" || config.SteamAPIKey == "null" {
		log.Fatal("STEAM_API_KEY is not set")
	}
	if config.ProxyUser == "" || config.ProxyPassword == "" || config.ProxyHost == "" || config.ProxyPort == "" {
		log.Fatal("Proxy configurations are not fully set")
	}
	if config.SeedSteamID == 0 {
		log.Fatal("SEED_STEAM_ID is not set or invalid")
	}

	InitDatabase(config.DatabasePath)
	defer db.Close()

	// Seed the database with the initial account
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	err = AddAccount(config.SeedSteamID, tx)
	if err != nil {
		tx.Rollback()
		log.Fatalf("Failed to add initial account: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	var ctx context.Context
	var cancel context.CancelFunc

	if config.RunDurationMinutes > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(config.RunDurationMinutes)*time.Minute)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	StartWorkers(ctx, config)

	// Wait for the context to be done
	<-ctx.Done()

	log.Println("Workers have finished, performing database checkpoint")

	// Perform a manual checkpoint to reset the WAL file
	_, err = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	if err != nil {
		log.Fatalf("Failed to checkpoint WAL: %v", err)
	}

	log.Println("Shutting down gracefully")
}
