package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr      string
	DatabaseDNS   string
	InternalToken string
	XUI           XUIConfig
}

type XUIConfig struct {
	BaseURL       string
	Username      string
	Password      string
	InboundID     int64
	ExternalHost  string
	Fingerprint   string
	SpiderX       string
	Flow          string
	HostHeader    string
	ServerName    string
	InsecureSkipVerify bool
}

func LoadConfig() Config {
	cfg := Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		DatabaseDNS:   mustEnv("DATABASE_URL"),
		InternalToken: mustEnv("INTERNAL_TOKEN"),
		XUI: XUIConfig{
			BaseURL:      mustEnv("XUI_BASE_URL"),
			Username:     mustEnv("XUI_USERNAME"),
			Password:     mustEnv("XUI_PASSWORD"),
			InboundID:    mustEnvInt64("XUI_INBOUND_ID"),
			ExternalHost: mustEnv("XUI_EXTERNAL_HOST"),
			Fingerprint:  getEnv("XUI_FINGERPRINT", "chrome"),
			SpiderX:      getEnv("XUI_SPIDERX", "/"),
			Flow:         getEnv("XUI_FLOW", ""),
			HostHeader:   getEnv("XUI_HOST_HEADER", ""),
			ServerName:   getEnv("XUI_SERVER_NAME", ""),
			InsecureSkipVerify: getEnv("XUI_INSECURE_SKIP_VERIFY", "") == "1",
		},
	}
	return cfg
}
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(key + " is required")
	}
	return v
}

func mustEnvInt64(key string) int64 {
	v := mustEnv(key)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic(key + " must be int64")
	}
	return n
}
