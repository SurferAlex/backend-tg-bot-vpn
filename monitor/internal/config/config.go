package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	VPNAPIBaseURL  string
	VPNAPIToken    string
	RedisURL       string
	XrayAccessLog  string
	AlertURL       string
	AlertToken     string
	PollInterval   time.Duration
	ActiveIPTTL    time.Duration
	AlertDedupTTL  time.Duration
	TargetsRefresh time.Duration
	StartAtEOF     bool
}

func Load() (Config, error) {
	base := strings.TrimSpace(os.Getenv("VPNAPI_BASE_URL"))
	if base == "" {
		return Config{}, fmt.Errorf("VPNAPI_BASE_URL is required")
	}
	token := strings.TrimSpace(os.Getenv("VPNAPI_INTERNAL_TOKEN"))
	if token == "" {
		return Config{}, fmt.Errorf("VPNAPI_INTERNAL_TOKEN is required")
	}
	logPath := strings.TrimSpace(os.Getenv("XRAY_ACCESS_LOG"))
	if logPath == "" {
		return Config{}, fmt.Errorf("XRAY_ACCESS_LOG is required")
	}

	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		redisURL = "redis://redis:6379/0"
	}

	poll := durationEnv("MONITOR_POLL_INTERVAL", 10*time.Second)
	activeTTL := durationEnv("MONITOR_ACTIVE_IP_TTL", 5*time.Minute)
	alertTTL := durationEnv("MONITOR_ALERT_DEDUP_TTL", durationEnv("MONITOR_BAN_DEDUP_TTL", 30*time.Minute))
	refresh := durationEnv("MONITOR_TARGETS_REFRESH", 45*time.Second)
	startEOF := os.Getenv("MONITOR_START_AT_EOF") != "0"

	return Config{
		VPNAPIBaseURL:  strings.TrimRight(base, "/"),
		VPNAPIToken:    token,
		RedisURL:       redisURL,
		XrayAccessLog:  logPath,
		AlertURL:       strings.TrimSpace(os.Getenv("ALERT_URL")),
		AlertToken:     strings.TrimSpace(os.Getenv("ALERT_INTERNAL_TOKEN")),
		PollInterval:   poll,
		ActiveIPTTL:    activeTTL,
		AlertDedupTTL:  alertTTL,
		TargetsRefresh: refresh,
		StartAtEOF:     startEOF,
	}, nil
}

func durationEnv(key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return def
}
