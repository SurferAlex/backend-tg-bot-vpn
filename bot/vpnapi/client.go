package vpnapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrAmbiguousClient = errors.New("ambiguous client name")

type Server struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateClientRequest struct {
	ServerID       string  `json:"serverId"`
	TelegramUserID *int64  `json:"telegramUserId,omitempty"`
	MaxIPs         int     `json:"maxIps"`
	TTLSeconds     int64   `json:"ttlSeconds"`
	Note           *string `json:"note,omitempty"`
}

type Client struct {
	ID             int64     `json:"id"`
	ClientUUID     string    `json:"clientUuid"`
	ServerID       string    `json:"serverId"`
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
			Timeout: 30 * time.Second,
		},
	}
}

func (a *API) ListServers(ctx context.Context) ([]Server, error) {
	u := a.baseURL + "/api/v1/servers"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Token", a.token)

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vpnapi list servers request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vpnapi list servers: status %d", resp.StatusCode)
	}

	var out struct {
		Servers []Server `json:"servers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("vpnapi list servers: decode response failed: %w", err)
	}
	return out.Servers, nil
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

func (a *API) ResolveClient(ctx context.Context, ref string) (Client, error) {
	u := a.baseURL + "/api/v1/clients/resolve?q=" + url.QueryEscape(strings.TrimSpace(ref))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Client{}, err
	}
	req.Header.Set("X-Internal-Token", a.token)
	resp, err := a.http.Do(req)
	if err != nil {
		return Client{}, fmt.Errorf("vpnapi resolve client request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return Client{}, ErrAmbiguousClient
	}
	if resp.StatusCode != http.StatusOK {
		return Client{}, fmt.Errorf("vpnapi resolve client: status %d", resp.StatusCode)
	}
	var out Client
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Client{}, fmt.Errorf("vpnapi resolve client: decode response failed: %w", err)
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

func (a *API) ExtendClient(ctx context.Context, clientUUID string, addDays int) (Client, error) {
	u := a.baseURL + "/api/v1/clients/" + url.PathEscape(clientUUID) + "/extend"
	body, err := json.Marshal(map[string]int{"addDays": addDays})
	if err != nil {
		return Client{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return Client{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", a.token)
	resp, err := a.http.Do(req)
	if err != nil {
		return Client{}, fmt.Errorf("vpnapi extend client request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Client{}, fmt.Errorf("vpnapi extend client: status %d", resp.StatusCode)
	}
	var out Client
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Client{}, fmt.Errorf("vpnapi extend client: decode response failed: %w", err)
	}
	return out, nil
}

func (a *API) UpdateMaxIPs(ctx context.Context, clientUUID string, maxIPs int) (Client, error) {
	u := a.baseURL + "/api/v1/clients/" + url.PathEscape(clientUUID) + "/max-ips"
	body, err := json.Marshal(map[string]int{"maxIps": maxIPs})
	if err != nil {
		return Client{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return Client{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", a.token)
	resp, err := a.http.Do(req)
	if err != nil {
		return Client{}, fmt.Errorf("vpnapi update max ips request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Client{}, fmt.Errorf("vpnapi update max ips: status %d", resp.StatusCode)
	}
	var out Client
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Client{}, fmt.Errorf("vpnapi update max ips: decode response failed: %w", err)
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
