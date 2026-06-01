package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	TelegramBotToken        string
	HappIOSAppStore         string
	HappRedirectPublicURL   string // https URL for inline «Открыть Happ», e.g. https://host/happ/open
	HappRedirectListenAddr  string // default :8091
	HappDefaultVless        string // fallback vless:// for «Открыть Happ» if user did not paste key
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
	return Config{
		TelegramBotToken:       token,
		HappIOSAppStore:        strings.TrimSpace(os.Getenv("HAPP_IOS_APP_STORE_URL")),
		HappRedirectPublicURL:  publicURL,
		HappRedirectListenAddr: strings.TrimSpace(os.Getenv("HAPP_REDIRECT_LISTEN")),
		HappDefaultVless:       strings.TrimSpace(os.Getenv("HAPP_DEFAULT_VLESS")),
	}, nil
}
