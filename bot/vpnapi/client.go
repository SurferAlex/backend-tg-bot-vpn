package vpnapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CreateClientRequest struct {
	TelegramUserID *int64  `json:"telegramUserId,omitempty"`
	MaxIPs         int     `json:"maxIps"`
	TTLSeconds     int64   `json:"ttlSeconds"`
	Note           *string `json:"note,omitempty"`
}

type Client struct {
	ID             int64     `json:"id"`
	ClientUUID     string    `json:"clientUuid"`
	TelegramUserID *int64    `json:"telegramUserId,omitempty"`
	MaxIPs         int       `json:"maxIps"`
	KeyExpiresAt   time.Time `json:"keyExpiresAt"`
	IsActive       bool      `json:"isActive"`
	Note           *string   `json:"note,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type API struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *API {
	return &API{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (a *API) CreateClient(ctx context.Context, req CreateClientRequest) (Client, error) {
	u := a.baseURL + "/api/v1/clients"

	body, err := json.Marshal(req)
	if err != nil {
		return Client{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return Client{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Token", a.token)

	resp, err := a.http.Do(httpReq)
	if err != nil {
		return Client{}, fmt.Errorf("vpnapi create client request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Client{}, fmt.Errorf("vpnapi create client: status %d", resp.StatusCode)
	}

	var out Client
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Client{}, fmt.Errorf("vpnapi create client: decode response failed: %w", err)
	}
	return out, nil
}

func (a *API) GetClient(ctx context.Context, clientUUID string) (Client, error) {
	u := a.baseURL + "/api/v1/clients/" + url.PathEscape(clientUUID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Client{}, err
	}
	req.Header.Set("X-Internal-Token", a.token)

	resp, err := a.http.Do(req)
	if err != nil {
		return Client{}, fmt.Errorf("vpnapi get client request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Client{}, fmt.Errorf("vpnapi get client: status %d", resp.StatusCode)
	}

	var out Client
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Client{}, fmt.Errorf("vpnapi get client: decode response failed: %w", err)
	}
	return out, nil
}

type Access struct {
	ClientUUID     string `json:"clientUuid"`
	InboundID      int64  `json:"inboundId"`
	XUIClientEmail string `json:"xuiClientEmail"`
	VLESSURI       string `json:"vlessUri"`
}

func (a *API) Provision(ctx context.Context, clientUUID string) (Access, error) {
	u := a.baseURL + "/api/v1/clients/" + url.PathEscape(clientUUID) + "/provision"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return Access{}, err
	}
	req.Header.Set("X-Internal-Token", a.token)
	resp, err := a.http.Do(req)
	if err != nil {
		return Access{}, fmt.Errorf("vpnapi provision request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Access{}, fmt.Errorf("vpnapi provision: status %d", resp.StatusCode)
	}
	var out Access
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Access{}, fmt.Errorf("vpnapi provision: decode response failed: %w", err)
	}
	return out, nil
}

func (a *API) GetAccess(ctx context.Context, clientUUID string) (Access, error) {
	u := a.baseURL + "/api/v1/clients/" + url.PathEscape(clientUUID) + "/access"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Access{}, err
	}
	req.Header.Set("X-Internal-Token", a.token)
	resp, err := a.http.Do(req)
	if err != nil {
		return Access{}, fmt.Errorf("vpnapi get access request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Access{}, fmt.Errorf("vpnapi get access: status %d", resp.StatusCode)
	}
	var out Access
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Access{}, fmt.Errorf("vpnapi get access: decode response failed: %w", err)
	}
	return out, nil
}

func (a *API) Revoke(ctx context.Context, clientUUID string) error {
	u := a.baseURL + "/api/v1/clients/" + url.PathEscape(clientUUID) + "/revoke"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Token", a.token)
	resp, err := a.http.Do(req)
	if err != nil {
		return fmt.Errorf("vpnapi revoke request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("vpnapi revoke: status %d", resp.StatusCode)
	}
	return nil
}
