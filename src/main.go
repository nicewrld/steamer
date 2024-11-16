package main

import (
	"log"
)

func main() {

	config := LoadConfig()

	// Validate essential configurations
	if config.SteamAPIKey == "" {
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

	StartWorkers(config)

	// Wait indefinitely to keep the application running
	select {}
}
