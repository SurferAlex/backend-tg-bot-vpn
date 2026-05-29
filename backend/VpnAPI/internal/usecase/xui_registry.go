package usecase

import (
	"api-vpn/internal/model"
	"api-vpn/internal/xui"
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"sync"
	"time"
)

type serverXUI struct {
	client   *xui.Client
	inbound  int64
	external string
	fp       string
	spiderX  string
	flow     string
}

// XUIRegistry routes X-UI API calls to the panel configured per VPN server.
type XUIRegistry struct {
	servers *VPNServers
	mu      sync.RWMutex
	cache   map[string]*serverXUI
}

func NewXUIRegistry(servers *VPNServers) *XUIRegistry {
	return &XUIRegistry{
		servers: servers,
		cache:   make(map[string]*serverXUI),
	}
}

func (r *XUIRegistry) forServer(ctx context.Context, serverID string) (*serverXUI, error) {
	serverID = strings.TrimSpace(serverID)
	if serverID == "" {
		return nil, ErrInvalidServer
	}

	r.mu.RLock()
	if sx, ok := r.cache[serverID]; ok {
		r.mu.RUnlock()
		return sx, nil
	}
	r.mu.RUnlock()

	srv, err := r.servers.GetActiveByID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	sx := buildServerXUI(srv)
	r.mu.Lock()
	r.cache[serverID] = sx
	r.mu.Unlock()
	return sx, nil
}

func buildServerXUI(srv model.VPNServer) *serverXUI {
	client := xui.New(srv.XUIBaseURL, srv.XUIUsername, srv.XUIPassword)
	if srv.XUIHostHeader != "" {
		client.WithHostHeader(srv.XUIHostHeader)
	}
	if strings.HasPrefix(strings.ToLower(srv.XUIBaseURL), "https://") &&
		(srv.XUIInsecureSkipVerify || srv.XUIServerName != "") {
		tlsCfg := &tls.Config{InsecureSkipVerify: srv.XUIInsecureSkipVerify}
		if srv.XUIServerName != "" {
			tlsCfg.ServerName = srv.XUIServerName
		}
		tr := &http.Transport{TLSClientConfig: tlsCfg}
		client.WithHTTPClient(&http.Client{Timeout: 10 * time.Second, Transport: tr})
	}
	return &serverXUI{
		client:   client,
		inbound:  srv.XUIInboundID,
		external: srv.XUIExternalHost,
		fp:       srv.XUIFingerprint,
		spiderX:  srv.XUISpiderX,
		flow:     srv.XUIFlow,
	}
}
