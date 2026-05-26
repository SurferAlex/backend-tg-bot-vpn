package vpnapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Target struct {
	ClientUUID     string `json:"clientUuid"`
	MaxIPs         int    `json:"maxIps"`
	XUIClientEmail string `json:"xuiClientEmail"`
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) ListMonitorTargets(ctx context.Context) ([]Target, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/monitor/targets", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vpnapi monitor targets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vpnapi monitor targets: status %d", resp.StatusCode)
	}

	var out struct {
		Targets []Target `json:"targets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("vpnapi monitor targets decode: %w", err)
	}
	return out.Targets, nil
}
