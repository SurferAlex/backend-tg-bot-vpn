package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr         string
	DatabaseDNS      string
	InternalToken    string
	XUI              XUIConfig
	OptionalServers  []OptionalVPNServer
}

// OptionalVPNServer is bootstrapped once (insert if missing) when env is set.
type OptionalVPNServer struct {
	ID   string
	Name string
	XUI  XUIConfig
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
		OptionalServers: loadOptionalVPNServers(),
	}
	return cfg
}

func loadOptionalVPNServers() []OptionalVPNServer {
	var out []OptionalVPNServer
	if s, ok := loadOptionalServerEnv("VPN_SERVER_VPS1", "vps_1", "VPS 1"); ok {
		out = append(out, s)
	}
	return out
}

func loadOptionalServerEnv(prefix, defaultID, defaultName string) (OptionalVPNServer, bool) {
	baseURL := strings.TrimSpace(os.Getenv(prefix + "_XUI_BASE_URL"))
	if baseURL == "" {
		return OptionalVPNServer{}, false
	}
	id := strings.TrimSpace(os.Getenv(prefix + "_ID"))
	if id == "" {
		id = defaultID
	}
	name := strings.TrimSpace(os.Getenv(prefix + "_NAME"))
	if name == "" {
		name = defaultName
	}
	return OptionalVPNServer{
		ID:   id,
		Name: name,
		XUI: XUIConfig{
			BaseURL:            baseURL,
			Username:           mustEnv(prefix + "_XUI_USERNAME"),
			Password:           mustEnv(prefix + "_XUI_PASSWORD"),
			InboundID:          mustEnvInt64(prefix + "_XUI_INBOUND_ID"),
			ExternalHost:       mustEnv(prefix + "_XUI_EXTERNAL_HOST"),
			Fingerprint:        getEnv(prefix+"_XUI_FINGERPRINT", "chrome"),
			SpiderX:            getEnv(prefix+"_XUI_SPIDERX", "/"),
			Flow:               getEnv(prefix+"_XUI_FLOW", ""),
			HostHeader:         getEnv(prefix+"_XUI_HOST_HEADER", ""),
			ServerName:         getEnv(prefix+"_XUI_SERVER_NAME", ""),
			InsecureSkipVerify: getEnv(prefix+"_XUI_INSECURE_SKIP_VERIFY", "") == "1",
		},
	}, true
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
