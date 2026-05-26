package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	url   string
	token string
	http  *http.Client
}

func New(url, token string) *Client {
	return &Client{
		url:   strings.TrimRight(strings.TrimSpace(url), "/"),
		token: strings.TrimSpace(token),
		http:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c.url != "" && c.token != ""
}

type Request struct {
	ClientUUID string `json:"client_uuid"`
	IPCount    int    `json:"ip_count"`
	MaxIPs     int    `json:"max_ips"`
	Text       string `json:"text,omitempty"`
}

func (c *Client) SendIPLimit(ctx context.Context, clientUUID string, ipCount, maxIPs int, excessIP string) error {
	if !c.Enabled() {
		return nil
	}
	text := fmt.Sprintf(
		"ALERT: превышение IP\nЛишний IP: %s\nИспользуется IP/лимит: %d/%d\nБан: 3x-ui + fail2ban",
		excessIP, ipCount, maxIPs,
	)
	body, err := json.Marshal(Request{
		ClientUUID: clientUUID,
		IPCount:    ipCount,
		MaxIPs:     maxIPs,
		Text:       text,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/internal/alert", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("alert: status %d", resp.StatusCode)
	}
	return nil
}
