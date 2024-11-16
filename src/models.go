package main

type Account struct {
	SteamID       int64
	Status        string
	LastUpdated   int64
	NumFriends    int
	IsPrivate     bool
	LockedBy      string
	LockTimestamp int64
}

type Friendship struct {
	AccountSteamID int64
	FriendSteamID  int64
	FriendSince    int64
}
