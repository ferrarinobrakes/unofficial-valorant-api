package valorant

import "fmt"

// riot response
type MatchHistoryResponse struct {
	Subject    string `json:"Subject"`
	BeginIndex int    `json:"BeginIndex"`
	EndIndex   int    `json:"EndIndex"`
	Total      int    `json:"Total"`
	History    []struct {
		MatchID       string `json:"MatchID"`
		GameStartTime int64  `json:"GameStartTime"`
		QueueID       string `json:"QueueID"`
	} `json:"History"`
}

func (c *Client) GetMatchHistory(shard, puuid, accessToken, entitlementToken string) (*MatchHistoryResponse, error) {
	url := fmt.Sprintf("https://pd.%s.a.pvp.net/match-history/v1/history/%s", shard, puuid)

	var result MatchHistoryResponse
	err := c.get(url, accessToken, entitlementToken, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get match history: %w", err)
	}

	return &result, nil
}
