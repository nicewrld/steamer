// src/worker.go
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

func StartWorkers(ctx context.Context, config Config) {
	var wg sync.WaitGroup

	for i := 0; i < config.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerID := generateWorkerID()
			RunWorker(ctx, workerID, config)
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
}

func generateWorkerID() string {
	return uuid.New().String()
}

func RunWorker(ctx context.Context, workerID string, config Config) {
	client := CreateHTTPClient(config)
	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			log.Printf("[%s] Shutting down worker", workerID)
			return
		default:
			// Proceed with processing
		}

		account, err := LockAccount(workerID, config.LockTimeout)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// No pending accounts; wait and try again
				select {
				case <-ctx.Done():
					log.Printf("[%s] Shutting down worker", workerID)
					return
				case <-time.After(1 * time.Second):
					continue
				}
			} else {
				log.Printf("[%s] Error locking account: %v", workerID, err)
				// Sleep and continue in case of transient errors
				select {
				case <-ctx.Done():
					log.Printf("[%s] Shutting down worker", workerID)
					return
				case <-time.After(1 * time.Second):
					continue
				}
			}
		}

		log.Printf("[%s] Processing account: %d", workerID, account.SteamID)

		friends, isPrivate, err := GetFriendList(account.SteamID, config.SteamAPIKey, client)
		if err != nil {
			log.Printf("[%s] Error fetching friends for account %d: %v", workerID, account.SteamID, err)
			// Update account as processed to avoid retrying indefinitely
			account.IsPrivate = false
			account.NumFriends = 0
			updateErr := UpdateAccountAfterProcessing(account)
			if updateErr != nil {
				log.Printf("[%s] Error updating account %d after processing: %v", workerID, account.SteamID, updateErr)
			}
			continue
		}

		account.IsPrivate = isPrivate

		if isPrivate {
			account.NumFriends = 0
		} else {
			account.NumFriends = len(friends)
			var friendships []Friendship
			for _, friend := range friends {
				friendSteamID, err := strconv.ParseInt(friend.SteamID, 10, 64)
				if err != nil {
					log.Printf("[%s] Invalid friend SteamID '%s': %v", workerID, friend.SteamID, err)
					continue
				}
				friendship := Friendship{
					AccountSteamID: account.SteamID,
					FriendSteamID:  friendSteamID,
					FriendSince:    friend.FriendSince,
				}
				friendships = append(friendships, friendship)
			}
			err = AddFriendships(account.SteamID, friendships)
			if err != nil {
				log.Printf("[%s] Error adding friendships for account %d: %v", workerID, account.SteamID, err)
			}
		}

		err = UpdateAccountAfterProcessing(account)
		if err != nil {
			log.Printf("[%s] Error updating account %d after processing: %v", workerID, account.SteamID, err)
		}
	}
}
