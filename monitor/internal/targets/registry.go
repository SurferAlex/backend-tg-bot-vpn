package targets

import (
	"strings"
	"sync"

	"vpn-monitor/internal/vpnapi"
)

type Entry struct {
	ClientUUID string
	MaxIPs     int
	Email      string
}

type Registry struct {
	mu     sync.RWMutex
	byEmail map[string]Entry
}

func NewRegistry() *Registry {
	return &Registry{byEmail: make(map[string]Entry)}
}

func (r *Registry) Replace(list []vpnapi.Target) {
	next := make(map[string]Entry, len(list))
	for _, t := range list {
		email := strings.TrimSpace(t.XUIClientEmail)
		if email == "" {
			continue
		}
		next[email] = Entry{
			ClientUUID: t.ClientUUID,
			MaxIPs:     t.MaxIPs,
			Email:      email,
		}
	}
	r.mu.Lock()
	r.byEmail = next
	r.mu.Unlock()
}

func (r *Registry) Lookup(email string) (Entry, bool) {
	email = strings.TrimSpace(email)
	r.mu.RLock()
	e, ok := r.byEmail[email]
	r.mu.RUnlock()
	return e, ok
}

func (r *Registry) Len() int {
	r.mu.RLock()
	n := len(r.byEmail)
	r.mu.RUnlock()
	return n
}
