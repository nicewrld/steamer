// src/steam_api.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SteamFriend struct {
	SteamID      string `json:"steamid"`
	Relationship string `json:"relationship"`
	FriendSince  int64  `json:"friend_since"`
}

type FriendsListResponse struct {
	FriendsList struct {
		Friends []SteamFriend `json:"friends"`
	} `json:"friendslist"`
}

func GetFriendList(steamID int64, apiKey string, client *http.Client) ([]SteamFriend, bool, error) {
	url := fmt.Sprintf(
		"http://api.steampowered.com/ISteamUser/GetFriendList/v1/?key=%s&steamid=%d&relationship=friend",
		apiKey, steamID,
	)

	resp, err := client.Get(url)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Account is private
		return nil, true, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("Steam API returned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	var friendsResponse FriendsListResponse
	err = json.Unmarshal(body, &friendsResponse)
	if err != nil {
		return nil, false, err
	}

	return friendsResponse.FriendsList.Friends, false, nil
}
