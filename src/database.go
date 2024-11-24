// src/database.go
package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDatabase(databasePath string) {
	var err error
	db, err = sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Enable Foreign Key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Failed to enable foreign key constraints: %v", err)
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}

	createTables()
}

func createTables() {
	accountsTable := `
	CREATE TABLE IF NOT EXISTS accounts (
		steamid INTEGER PRIMARY KEY,
		status TEXT DEFAULT 'pending',
		last_updated INTEGER,
		num_friends INTEGER,
		is_private BOOLEAN DEFAULT 0,
		locked_by TEXT,
		lock_timestamp INTEGER
	);`

	friendshipsTable := `
	CREATE TABLE IF NOT EXISTS friendships (
		account_steamid INTEGER,
		friend_steamid INTEGER,
		friend_since INTEGER,
		PRIMARY KEY (account_steamid, friend_steamid),
		FOREIGN KEY (account_steamid) REFERENCES accounts(steamid),
		FOREIGN KEY (friend_steamid) REFERENCES accounts(steamid)
	);`

	_, err := db.Exec(accountsTable)
	if err != nil {
		log.Fatalf("Failed to create accounts table: %v", err)
	}

	_, err = db.Exec(friendshipsTable)
	if err != nil {
		log.Fatalf("Failed to create friendships table: %v", err)
	}

	createIndexes()
}

func createIndexes() {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);",
		"CREATE INDEX IF NOT EXISTS idx_accounts_locked_by ON accounts(locked_by);",
		"CREATE INDEX IF NOT EXISTS idx_friendships_account_steamid ON friendships(account_steamid);",
	}

	for _, idxQuery := range indexes {
		_, err := db.Exec(idxQuery)
		if err != nil {
			log.Fatalf("Failed to create index: %v", err)
		}
	}
}

func AddAccount(steamID int64, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT OR IGNORE INTO accounts (steamid) VALUES (?)", steamID)
	return err
}

func LockAccount(workerID string, lockTimeout int64) (*Account, error) {
	var account Account
	now := time.Now().Unix()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// Clean up stale locks
	_, err = tx.Exec(`
		UPDATE accounts
		SET status = 'pending', locked_by = NULL, lock_timestamp = NULL
		WHERE status = 'processing' AND lock_timestamp < ?;
	`, now-lockTimeout)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Atomically select and lock an account
	row := tx.QueryRow(`
		SELECT steamid FROM accounts
		WHERE status = 'pending' AND (locked_by IS NULL OR lock_timestamp < ?)
		LIMIT 1
	`, now-lockTimeout)

	err = row.Scan(&account.SteamID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE accounts
		SET status = 'processing', locked_by = ?, lock_timestamp = ?
		WHERE steamid = ?;
	`, workerID, now, account.SteamID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	account.Status = "processing"
	account.LockedBy = workerID
	account.LockTimestamp = now

	return &account, nil
}

func UpdateAccountAfterProcessing(account *Account) error {
	_, err := db.Exec(`
		UPDATE accounts
		SET status = 'processed', locked_by = NULL, lock_timestamp = NULL,
			last_updated = ?, num_friends = ?, is_private = ?
		WHERE steamid = ? AND locked_by = ?;
	`, time.Now().Unix(), account.NumFriends, account.IsPrivate, account.SteamID, account.LockedBy)
	return err
}

func AddFriendships(accountSteamID int64, friendships []Friendship) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmtFriendship, err := tx.Prepare(`
		INSERT OR IGNORE INTO friendships (account_steamid, friend_steamid, friend_since)
		VALUES (?, ?, ?);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmtFriendship.Close()

	stmtAccount, err := tx.Prepare(`
		INSERT OR IGNORE INTO accounts (steamid)
		VALUES (?);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmtAccount.Close()

	for _, friendship := range friendships {
		_, err = stmtFriendship.Exec(friendship.AccountSteamID, friendship.FriendSteamID, friendship.FriendSince)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Add the friend account within the same transaction
		err = AddAccount(friendship.FriendSteamID, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}
