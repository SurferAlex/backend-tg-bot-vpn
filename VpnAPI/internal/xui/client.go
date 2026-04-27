package xui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	username string
	password string

	http   *http.Client
	cookie string

	hostHeader string
}

func New(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	if httpClient != nil {
		c.http = httpClient
	}
	return c
}

func (c *Client) WithHostHeader(host string) *Client {
	c.hostHeader = strings.TrimSpace(host)
	return c
}

type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Obj     T      `json:"obj"`
}

func (c *Client) login(ctx context.Context) error {
	u := c.baseURL + "/login"
	form := url.Values{}
	form.Set("username", c.username)
	form.Set("password", c.password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var out apiResponse[any]
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != http.StatusOK || !out.Success {
		return fmt.Errorf("xui login failed: status=%d msg=%s", resp.StatusCode, out.Msg)
	}

	for _, ck := range resp.Cookies() {
		if ck.Name == "3x-ui" {
			c.cookie = ck.Name + "=" + ck.Value
			return nil
		}
	}
	if sc := resp.Header.Get("Set-Cookie"); sc != "" {
		parts := strings.Split(sc, ";")
		if len(parts) > 0 && strings.Contains(parts[0], "=") {
			c.cookie = strings.TrimSpace(parts[0])
			return nil
		}
	}
	return fmt.Errorf("xui login: cookie not found")
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any) error {
	if c.cookie == "" {
		if err := c.login(ctx); err != nil {
			return err
		}
	}

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	u := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Cookie", c.cookie)
	if c.hostHeader != "" {
		req.Host = c.hostHeader
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.cookie = ""
		return fmt.Errorf("xui unauthorized")
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

type inboundObj struct {
	ID             int64  `json:"id"`
	Port           int64  `json:"port"`
	Protocol       string `json:"protocol"`
	Remark         string `json:"remark"`
	StreamSettings string `json:"streamSettings"`
	Settings       string `json:"settings"`
}

type streamSettings struct {
	Security        string `json:"security"`
	Network         string `json:"network"`
	RealitySettings struct {
		ServerNames []string `json:"serverNames"`
		ShortIds    []string `json:"shortIds"`
		Settings    struct {
			PublicKey string `json:"publicKey"`
		} `json:"settings"`
		SpiderX string `json:"spiderX"`
	} `json:"realitySettings"`
}

func (c *Client) GetInbound(ctx context.Context, inboundID int64) (inboundObj, streamSettings, error) {
	var resp apiResponse[inboundObj]
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/panel/api/inbounds/get/%d", inboundID), nil, &resp); err != nil {
		return inboundObj{}, streamSettings{}, err
	}
	if !resp.Success {
		return inboundObj{}, streamSettings{}, fmt.Errorf("xui get inbound: %s", resp.Msg)
	}
	var ss streamSettings
	if err := json.Unmarshal([]byte(resp.Obj.StreamSettings), &ss); err != nil {
		return resp.Obj, streamSettings{}, fmt.Errorf("parse streamSettings: %w", err)
	}
	return resp.Obj, ss, nil
}

type addClientReq struct {
	ID       int64  `json:"id"`
	Settings string `json:"settings"`
}

type xrayClient struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Enable     bool   `json:"enable"`
	LimitIP    int    `json:"limitIp"`
	TotalGB    int64  `json:"totalGB"`
	ExpiryTime int64  `json:"expiryTime"`
	Flow       string `json:"flow"`
}

func (c *Client) AddOrUpdateVLESSClient(ctx context.Context, inboundID int64, clientUUID string, email string, limitIP int, expiryTimeMs int64, flow string) error {
	settings := map[string]any{
		"clients": []xrayClient{{
			ID:         clientUUID,
			Email:      email,
			Enable:     true,
			LimitIP:    limitIP,
			TotalGB:    0,
			ExpiryTime: expiryTimeMs,
			Flow:       flow,
		}},
	}
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	var resp apiResponse[any]
	req := addClientReq{ID: inboundID, Settings: string(b)}
	if err := c.doJSON(ctx, http.MethodPost, "/panel/api/inbounds/addClient", req, &resp); err != nil {
		return err
	}
	if resp.Success {
		return nil
	}

	up := map[string]any{"id": inboundID, "settings": string(b)}
	if err := c.doJSON(ctx, http.MethodPost, "/panel/api/inbounds/updateClient/"+url.PathEscape(clientUUID), up, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("xui add/update client: %s", resp.Msg)
	}
	return nil
}

func (c *Client) DeleteClientByEmail(ctx context.Context, inboundID int64, email string) error {
	var resp apiResponse[any]
	path := fmt.Sprintf("/panel/api/inbounds/%d/delClientByEmail/%s", inboundID, url.PathEscape(email))
	if err := c.doJSON(ctx, http.MethodPost, path, nil, &resp); err != nil {
		return err
	}
	if !resp.Success {

		return fmt.Errorf("xui delete client: %s", resp.Msg)
	}
	return nil
}

func BuildVLESSRealityURI(externalHost string, port int64, userUUID string, label string, ss streamSettings, fp string, spiderX string, flow string) (string, error) {
	if externalHost == "" {
		return "", fmt.Errorf("external host is empty")
	}
	if ss.RealitySettings.Settings.PublicKey == "" {
		return "", fmt.Errorf("reality publicKey is empty")
	}
	if len(ss.RealitySettings.ServerNames) == 0 {
		return "", fmt.Errorf("reality serverNames is empty")
	}
	if len(ss.RealitySettings.ShortIds) == 0 {
		return "", fmt.Errorf("reality shortIds is empty")
	}
	sni := ss.RealitySettings.ServerNames[0]
	sid := ss.RealitySettings.ShortIds[0]
	if fp == "" {
		fp = "chrome"
	}
	if spiderX == "" {
		spiderX = "/"
	}
	q := url.Values{}
	q.Set("type", "tcp")
	q.Set("security", "reality")
	q.Set("pbk", ss.RealitySettings.Settings.PublicKey)
	q.Set("fp", fp)
	q.Set("sni", sni)
	q.Set("sid", sid)
	q.Set("spx", spiderX)
	if flow != "" {
		q.Set("flow", flow)
	}

	return fmt.Sprintf("vless://%s@%s:%d?%s#%s", userUUID, externalHost, port, q.Encode(), url.QueryEscape(label)), nil
}
