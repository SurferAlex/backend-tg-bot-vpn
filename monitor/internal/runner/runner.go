package runner

import (
	"context"
	"log"
	"time"

	"vpn-monitor/internal/alert"
	"vpn-monitor/internal/config"
	"vpn-monitor/internal/limit"
	"vpn-monitor/internal/logtail"
	"vpn-monitor/internal/parser"
	"vpn-monitor/internal/redisstore"
	"vpn-monitor/internal/targets"
	"vpn-monitor/internal/vpnapi"

	"github.com/redis/go-redis/v9"
)

type Runner struct {
	cfg      config.Config
	vpn      *vpnapi.Client
	registry *targets.Registry
	enforcer *limit.Enforcer
	alerts   *alert.Client
}

func New(cfg config.Config, vpn *vpnapi.Client, rdb *redis.Client) *Runner {
	reg := targets.NewRegistry()
	store := redisstore.New(rdb, cfg.ActiveIPTTL, cfg.AlertDedupTTL)

	return &Runner{
		cfg:      cfg,
		vpn:      vpn,
		registry: reg,
		enforcer: limit.New(store, reg),
		alerts:   alert.New(cfg.AlertURL, cfg.AlertToken),
	}
}

func (r *Runner) Run(ctx context.Context) error {
	if err := r.refreshTargets(ctx); err != nil {
		log.Printf("monitor: initial targets refresh failed: %v", err)
	}

	refreshTicker := time.NewTicker(r.cfg.TargetsRefresh)
	defer refreshTicker.Stop()

	pollTicker := time.NewTicker(r.cfg.PollInterval)
	defer pollTicker.Stop()

	var tail *logtail.Tailer

	for {
		if tail == nil {
			t, err := logtail.Open(r.cfg.XrayAccessLog, r.cfg.StartAtEOF)
			if err != nil {
				log.Printf("monitor: waiting for access log %s: %v", r.cfg.XrayAccessLog, err)
			} else {
				tail = t
				log.Printf("monitor: tailing %s (start_at_eof=%v)", r.cfg.XrayAccessLog, r.cfg.StartAtEOF)
			}
		}

		select {
		case <-ctx.Done():
			if tail != nil {
				_ = tail.Close()
			}
			return ctx.Err()
		case <-refreshTicker.C:
			if err := r.refreshTargets(ctx); err != nil {
				log.Printf("monitor: targets refresh failed: %v", err)
			}
		case <-pollTicker.C:
			if tail == nil {
				t, err := logtail.Open(r.cfg.XrayAccessLog, r.cfg.StartAtEOF)
				if err != nil {
					continue
				}
				tail = t
				log.Printf("monitor: tailing %s", r.cfg.XrayAccessLog)
			}
			if err := r.processTail(ctx, tail); err != nil {
				_ = tail.Close()
				tail = nil
			}
		}
	}
}

func (r *Runner) refreshTargets(ctx context.Context) error {
	list, err := r.vpn.ListMonitorTargets(ctx)
	if err != nil {
		return err
	}
	r.registry.Replace(list)
	log.Printf("monitor: loaded %d targets", r.registry.Len())
	return nil
}

func (r *Runner) processTail(ctx context.Context, tail *logtail.Tailer) error {
	lines, err := tail.ReadLines()
	if err != nil {
		log.Printf("monitor: read log failed: %v", err)
		return err
	}
	if len(lines) == 0 {
		return nil
	}
	now := time.Now()
	for _, line := range lines {
		email, ip, ok := parser.ParseAccessLine(line)
		if !ok {
			continue
		}
		ev, err := r.enforcer.HandleConnection(ctx, email, ip, now)
		if err != nil {
			log.Printf("monitor: handle connection failed (email=%s ip=%s): %v", email, ip, err)
			continue
		}
		if ev == nil {
			continue
		}
		log.Printf(
			"monitor: IP limit exceeded client_uuid=%s email=%s excess_ip=%s ip_count=%d max_ips=%d",
			ev.ClientUUID, ev.Email, ev.IP, ev.IPCount, ev.MaxIPs,
		)
		if err := r.alerts.SendIPLimit(ctx, ev.ClientUUID, ev.IPCount, ev.MaxIPs, ev.IP); err != nil {
			log.Printf("monitor: alert failed (client_uuid=%s): %v", ev.ClientUUID, err)
		}
	}
	return nil
}
