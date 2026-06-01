package happ

import (
	"net/url"
	"strings"
)

const (
	DeeplinkScheme = "happ://"

	DefaultIOSAppStore = "https://apps.apple.com/ru/app/happ-proxy-utility-plus/id6746188973"

	DeeplinkAddPrefix = "happ://add/"

	// DefaultOpenRedirectPublicURL used when HAPP_REDIRECT_PUBLIC_URL is not set (nginx /happ/ → :8091/open).
	DefaultOpenRedirectPublicURL = "https://panel.alexsurfervpn.space/happ/open"
)

// AddConfigURL builds happ://add/… for vless:// or https:// subscription links.
func AddConfigURL(config string) string {
	config = strings.TrimSpace(config)
	if config == "" {
		return DeeplinkScheme
	}
	lower := strings.ToLower(config)
	if strings.HasPrefix(lower, "vless://") ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") {
		return DeeplinkAddPrefix + config
	}
	return DeeplinkAddPrefix + config
}

// OpenAppURL opens Happ with optional config (vless:// or https:// subscription).
func OpenAppURL(config string) string {
	return AddConfigURL(config)
}

// OpenRedirectURL is the https link for Telegram inline «Открыть Happ» (redirects to happ://add/…).
func OpenRedirectURL(publicBase, vless string) string {
	base := strings.TrimSuffix(strings.TrimSpace(publicBase), "/")
	if base == "" || !strings.HasPrefix(base, "https://") {
		return ""
	}
	vless = strings.TrimSpace(vless)
	if vless == "" {
		return base
	}
	return base + "?vless=" + url.QueryEscape(vless)
}

// ResolveOpenRedirectBase returns env override or default public redirect URL.
func ResolveOpenRedirectBase(override string) string {
	if s := strings.TrimSpace(override); s != "" {
		return strings.TrimSuffix(s, "/")
	}
	return DefaultOpenRedirectPublicURL
}

func AppStoreURL(storeOverride string) string {
	if s := strings.TrimSpace(storeOverride); s != "" {
		return s
	}
	return DefaultIOSAppStore
}
