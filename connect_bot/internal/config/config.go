package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	TelegramBotToken        string
	HappIOSAppStore         string
	HappRedirectPublicURL   string
	HappRedirectListenAddr  string
	HappDefaultRoutingB64   string
	HappRoutingOnAdd        bool // happ://routing/onadd/ vs /add/
}

func Load() (Config, error) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	publicURL := strings.TrimSpace(os.Getenv("HAPP_REDIRECT_PUBLIC_URL"))
	if publicURL != "" && !strings.HasPrefix(publicURL, "https://") {
		return Config{}, fmt.Errorf("HAPP_REDIRECT_PUBLIC_URL must start with https://")
	}
	onAdd := true
	if v := strings.TrimSpace(os.Getenv("HAPP_ROUTING_ONADD")); v != "" {
		onAdd = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}
	return Config{
		TelegramBotToken:       token,
		HappIOSAppStore:        strings.TrimSpace(os.Getenv("HAPP_IOS_APP_STORE_URL")),
		HappRedirectPublicURL:  publicURL,
		HappRedirectListenAddr: strings.TrimSpace(os.Getenv("HAPP_REDIRECT_LISTEN")),
		HappDefaultRoutingB64:  strings.TrimSpace(os.Getenv("HAPP_DEFAULT_ROUTING_B64")),
		HappRoutingOnAdd:       onAdd,
	}, nil
}
