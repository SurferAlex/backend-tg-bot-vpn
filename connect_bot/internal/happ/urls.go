package happ

import (
	"net/url"
	"strings"
)

const (
	DeeplinkScheme = "happ://"

	DefaultIOSAppStore = "https://apps.apple.com/ru/app/happ-proxy-utility-plus/id6746188973"

	DeeplinkAddPrefix = "happ://add/"

	DefaultOpenRedirectPublicURL = "https://sub.alexsurfervpn.space/happ/open"
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

// OpenRedirectURLVless builds https link with ?vless= for VPN key import.
func OpenRedirectURLVless(publicBase, vless string) string {
	base := strings.TrimSuffix(strings.TrimSpace(publicBase), "/")
	vless = strings.TrimSpace(vless)
	if base == "" || !strings.HasPrefix(base, "https://") || vless == "" {
		return ""
	}
	return base + "?vless=" + url.QueryEscape(vless)
}

// OpenRedirectURL builds https link with ?routing= for routing profile.
func OpenRedirectURL(publicBase, routingB64 string) string {
	base := strings.TrimSuffix(strings.TrimSpace(publicBase), "/")
	if base == "" || !strings.HasPrefix(base, "https://") {
		return ""
	}
	routingB64 = strings.TrimSpace(routingB64)
	if routingB64 == "" {
		routingB64 = DefaultRoutingProfileB64
	}
	return base + "?routing=" + url.QueryEscape(routingB64)
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
