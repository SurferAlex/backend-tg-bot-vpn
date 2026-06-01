package happ

import "strings"

// OpenTargetKind is the deeplink type for /happ/open redirect.
type OpenTargetKind int

const (
	OpenTargetRouting OpenTargetKind = iota
	OpenTargetVless
)

// ResolveOpenTarget builds happ:// deeplink from query params (vless wins over routing).
func ResolveOpenTarget(vless, routingB64 string, routingAutoOnAdd bool, defaultRoutingB64 string) (target string, kind OpenTargetKind) {
	vless = strings.TrimSpace(vless)
	if vless != "" {
		return AddConfigURL(vless), OpenTargetVless
	}
	b64 := ResolveRoutingB64(routingB64, defaultRoutingB64)
	return OpenRoutingTarget(b64, routingAutoOnAdd), OpenTargetRouting
}
