package lcu

type EntitlementsTokenResponse struct {
	AccessToken  string   `json:"accessToken"`
	Entitlements []string `json:"entitlements"`
	Issuer       string   `json:"issuer"`
	Subject      string   `json:"subject"`
	Token        string   `json:"token"`
}

func (c *Client) GetEntitlementsToken() (*EntitlementsTokenResponse, error) {
	var result EntitlementsTokenResponse
	err := c.get("/entitlements/v1/token", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
