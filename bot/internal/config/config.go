package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TelegramConfig struct {
	BotToken string
	AdminIDs []int64
}

type AlertConfig struct {
	HTTPAddr      string
	InternalToken string
}

type VPNAPIConfig struct {
	BaseURL       string
	InternalToken string
}

type Config struct {
	Telegram TelegramConfig
	Alert    AlertConfig
	VPNAPI   VPNAPIConfig
}

func Load() (Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	raw := strings.TrimSpace(os.Getenv("ADMIN_TELEGRAM_USER_ID"))
	if raw == "" {
		return Config{}, fmt.Errorf("ADMIN_TELEGRAM_USER_ID is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("ADMIN_TELEGRAM_USER_ID: %w", err)
	}

	alertAddr := strings.TrimSpace(os.Getenv("ALERT_HTTP_ADDR"))
	if alertAddr == "" {
		alertAddr = ":8090"
	}
	alertToken := strings.TrimSpace(os.Getenv("ALERT_INTERNAL_TOKEN"))
	if alertToken == "" {
		return Config{}, fmt.Errorf("ALERT_INTERNAL_TOKEN is required")
	}

	vpnBaseURL := strings.TrimSpace(os.Getenv("VPNAPI_BASE_URL"))
	if vpnBaseURL == "" {
		return Config{}, fmt.Errorf("VPNAPI_BASE_URL is required")
	}
	vpnToken := strings.TrimSpace(os.Getenv("VPNAPI_INTERNAL_TOKEN"))
	if vpnToken == "" {
		return Config{}, fmt.Errorf("VPNAPI_INTERNAL_TOKEN is required")
	}

	return Config{
		Telegram: TelegramConfig{BotToken: token, AdminIDs: []int64{id}},
		Alert:    AlertConfig{HTTPAddr: alertAddr, InternalToken: alertToken},
		VPNAPI:   VPNAPIConfig{BaseURL: vpnBaseURL, InternalToken: vpnToken},
	}, nil
}
