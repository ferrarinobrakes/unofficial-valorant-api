package valorant

import "fmt"

type Player struct {
	Subject      string `json:"subject"`
	GameName     string `json:"gameName"`
	TagLine      string `json:"tagLine"`
	TeamID       string `json:"teamId"`
	CharacterID  string `json:"characterId"`
	PlayerCard   string `json:"playerCard"`
	PlayerTitle  string `json:"playerTitle"`
	AccountLevel int    `json:"accountLevel"`
}

type MatchDetailsResponse struct {
	MatchInfo struct {
		MatchID string `json:"matchId"`
	} `json:"matchInfo"`
	Players []Player `json:"players"`
}

func (c *Client) GetMatchDetails(shard, matchID, accessToken, entitlementToken string) (*MatchDetailsResponse, error) {
	url := fmt.Sprintf("https://pd.%s.a.pvp.net/match-details/v1/matches/%s", shard, matchID)

	var result MatchDetailsResponse
	err := c.get(url, accessToken, entitlementToken, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get match details: %w", err)
	}

	return &result, nil
}

func (m *MatchDetailsResponse) FindPlayerByPUUID(puuid string) *Player {
	for i := range m.Players {
		if m.Players[i].Subject == puuid {
			return &m.Players[i]
		}
	}
	return nil
}
