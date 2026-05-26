package limit

import (
	"context"
	"time"

	"vpn-monitor/internal/redisstore"
	"vpn-monitor/internal/targets"
)

type Enforcer struct {
	store     *redisstore.Store
	registry  *targets.Registry
}

func New(store *redisstore.Store, registry *targets.Registry) *Enforcer {
	return &Enforcer{
		store:     store,
		registry:  registry,
	}
}

// LimitEvent is emitted when active IP count exceeds max_ips for a monitored client.
type LimitEvent struct {
	ClientUUID string
	Email      string
	IP         string // newest / excess IP at detection time
	IPCount    int
	MaxIPs     int
}

func (e *Enforcer) HandleConnection(ctx context.Context, email, ip string, at time.Time) (*LimitEvent, error) {
	entry, ok := e.registry.Lookup(email)
	if !ok {
		return nil, nil
	}
	if entry.MaxIPs <= 0 {
		return nil, nil
	}

	if err := e.store.RecordConnection(ctx, entry.ClientUUID, ip, at); err != nil {
		return nil, err
	}

	active, err := e.store.ListActiveIPs(ctx, entry.ClientUUID)
	if err != nil {
		return nil, err
	}
	if len(active) <= entry.MaxIPs {
		return nil, nil
	}

	newestIP := ip
	var newestTS int64
	for _, a := range active {
		ts := a.LastSeen.Unix()
		if ts >= newestTS {
			newestTS = ts
			newestIP = a.IP
		}
	}

	alerted, err := e.store.IsAlerted(ctx, entry.ClientUUID, newestIP)
	if err != nil {
		return nil, err
	}
	if alerted {
		return nil, nil
	}

	if err := e.store.MarkAlerted(ctx, entry.ClientUUID, newestIP); err != nil {
		return nil, err
	}

	return &LimitEvent{
		ClientUUID: entry.ClientUUID,
		Email:      entry.Email,
		IP:         newestIP,
		IPCount:    len(active),
		MaxIPs:     entry.MaxIPs,
	}, nil
}
