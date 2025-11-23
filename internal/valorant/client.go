package valorant

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const ClientPlatform = "ew0KCSJwbGF0Zm9ybVR5cGUiOiAiUEMiLA0KCSJwbGF0Zm9ybU9TIjogIldpbmRvd3MiLA0KCSJwbGF0Zm9ybU9TVmVyc2lvbiI6ICIxMC4wLjE5MDQyLjEuMjU2LjY0Yml0IiwNCgkicGxhdGZvcm1DaGlwc2V0IjogIlVua25vd24iDQp9"

type Client struct {
	httpClient *http.Client
	logger     *zap.SugaredLogger
}

func NewClient(logger *zap.SugaredLogger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
		logger: logger,
	}
}

func (c *Client) doRequest(method, url string, accessToken, entitlementToken string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Riot-ClientPlatform", ClientPlatform)
	req.Header.Set("X-Riot-ClientVersion", "unknown")
	req.Header.Set("X-Riot-Entitlements-JWT", entitlementToken)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	c.logger.Debugw("Riot API request", "method", method, "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	c.logger.Debugw("Riot API response", "status", resp.StatusCode)

	return resp, nil
}

func (c *Client) get(url string, accessToken, entitlementToken string, result interface{}) error {
	resp, err := c.doRequest("GET", url, accessToken, entitlementToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func RegionToShard(region string) string {
	switch region {
	case "na1", "na2", "na3", "latam", "br":
		return "na"
	case "pbe":
		return "pbe"
	case "eu1", "eu2", "eu3":
		return "eu"
	case "ap1", "ap2", "ap3":
		return "ap"
	case "kr1":
		return "kr"
	case "sa", "sa1", "sa2":
		return "eu"
	default:
		if len(region) >= 2 {
			prefix := region[:2]
			switch prefix {
			case "na", "la", "br":
				return "na"
			case "eu":
				return "eu"
			case "ap":
				return "ap"
			case "kr":
				return "kr"
			}
		}
		return "na"
	}
}
