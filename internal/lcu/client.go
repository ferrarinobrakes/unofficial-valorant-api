package lcu

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
	logger     *zap.SugaredLogger
}

func NewClient(lockfile *LockfileData, logger *zap.SugaredLogger) *Client {
	return &Client{
		baseURL:    lockfile.GetBaseURL(),
		authHeader: lockfile.GetAuthHeader(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // no need
				},
			},
		},
		logger: logger,
	}
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")

	c.logger.Debugw("LCU request", "method", method, "path", path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	c.logger.Debugw("LCU response", "status", resp.StatusCode)

	return resp, nil
}

func (c *Client) get(path string, result interface{}) error {
	resp, err := c.doRequest("GET", path, nil)
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

func (c *Client) post(path string, body io.Reader, result interface{}) error {
	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

func (c *Client) delete(path string, body io.Reader) error {
	resp, err := c.doRequest("DELETE", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
