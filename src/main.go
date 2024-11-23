package main

import (
	"context"
	"log"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {

	err := profiler.Start(
		profiler.WithService("<SERVICE_NAME>"),
		profiler.WithEnv("<ENVIRONMENT>"),
		profiler.WithVersion("<APPLICATION_VERSION>"),
		profiler.WithTags("<KEY1>:<VALUE1>", "<KEY2>:<VALUE2>"),
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

	log.Println("Shutting down gracefully")
}
