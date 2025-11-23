package lcu

import (
	"bytes"
	"encoding/json"
)

type FriendRequest struct {
	GameName     string `json:"game_name"`
	GameTag      string `json:"game_tag"`
	Name         string `json:"name"`
	Note         string `json:"note"`
	PID          string `json:"pid"`
	Platform     string `json:"platform"`
	PUUID        string `json:"puuid"`
	Region       string `json:"region"`
	Subscription string `json:"subscription"`
}

type FriendRequestsResponse struct {
	Requests []FriendRequest `json:"requests"`
}

type SendFriendRequestBody struct {
	GameName string `json:"game_name"`
	GameTag  string `json:"game_tag"`
}

type RemoveFriendRequestBody struct {
	PUUID string `json:"puuid"`
}

func (c *Client) SendFriendRequest(gameName, gameTag string) error {
	body := SendFriendRequestBody{
		GameName: gameName,
		GameTag:  gameTag,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.post("/chat/v4/friendrequests", bytes.NewReader(jsonBody), nil)
}

func (c *Client) GetFriendRequests() (*FriendRequestsResponse, error) {
	var result FriendRequestsResponse
	err := c.get("/chat/v3/friendrequests", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteFriendRequest(puuid string) error {
	body := RemoveFriendRequestBody{
		PUUID: puuid,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.delete("/chat/v4/friendrequests", bytes.NewReader(jsonBody))
}
