package state

import (
	"strings"
	"sync"
	"time"

	"vpn-monitor/internal/vpnapi"
)

type clientTarget struct {
	clientUUID string
	maxIPs     int
}

type Tracker struct {
	window time.Duration
	dedup  time.Duration

	mu       sync.Mutex
	byEmail  map[string]clientTarget
	ips      map[string]map[string]time.Time
	lastViol map[string]time.Time
}

func NewTracker(window, dedup time.Duration) *Tracker {
	return &Tracker{
		window:   window,
		dedup:    dedup,
		byEmail:  make(map[string]clientTarget),
		ips:      make(map[string]map[string]time.Time),
		lastViol: make(map[string]time.Time),
	}
}

func (t *Tracker) SetTargets(targets []vpnapi.Target) {
	t.mu.Lock()
	defer t.mu.Unlock()

	next := make(map[string]clientTarget, len(targets))
	for _, tg := range targets {
		email := strings.TrimSpace(tg.XUIClientEmail)
		if email == "" {
			continue
		}
		next[email] = clientTarget{
			clientUUID: tg.ClientUUID,
			maxIPs:     tg.MaxIPs,
		}
	}
	t.byEmail = next

	for email := range t.ips {
		if _, ok := next[email]; !ok {
			delete(t.ips, email)
		}
	}
}

type Violation struct {
	ClientUUID string
	Email      string
	IPCount    int
	MaxIPs     int
	IPs        []string
}

func (t *Tracker) Record(email, ip string, now time.Time) (Violation, bool) {
	email = strings.TrimSpace(email)
	if email == "" || ip == "" {
		return Violation{}, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	target, ok := t.byEmail[email]
	if !ok {
		return Violation{}, false
	}

	m, ok := t.ips[email]
	if !ok {
		m = make(map[string]time.Time)
		t.ips[email] = m
	}
	m[ip] = now

	cutoff := now.Add(-t.window)
	for addr, seen := range m {
		if seen.Before(cutoff) {
			delete(m, addr)
		}
	}

	ipCount := len(m)
	if ipCount <= target.maxIPs {
		return Violation{}, false
	}

	if last, ok := t.lastViol[target.clientUUID]; ok && now.Sub(last) < t.dedup {
		return Violation{}, false
	}
	t.lastViol[target.clientUUID] = now

	ips := make([]string, 0, ipCount)
	for addr := range m {
		ips = append(ips, addr)
	}

	return Violation{
		ClientUUID: target.clientUUID,
		Email:      email,
		IPCount:    ipCount,
		MaxIPs:     target.maxIPs,
		IPs:        ips,
	}, true
}
