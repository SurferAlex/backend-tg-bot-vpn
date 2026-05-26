package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"vpn-monitor/internal/config"
	"vpn-monitor/internal/runner"
	"vpn-monitor/internal/vpnapi"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("monitor: redis url: %v", err)
	}
	rdb := redis.NewClient(opt)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("monitor: redis ping: %v", err)
	}

	vpn := vpnapi.New(cfg.VPNAPIBaseURL, cfg.VPNAPIToken)
	r := runner.New(cfg, vpn, rdb)

	log.Printf("monitor: started mode=alert_only poll=%s active_ttl=%s alert_dedup_ttl=%s",
		cfg.PollInterval, cfg.ActiveIPTTL, cfg.AlertDedupTTL)

	if err := r.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
