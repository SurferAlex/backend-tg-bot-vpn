package happ

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	RoutingAddPrefix   = "happ://routing/add/"
	RoutingOnAddPrefix = "happ://routing/onadd/"

	// Default test profile from Happ docs (https://www.happ.su/main/dev-docs/routing).
	DefaultRoutingProfileB64 = "ewogICAgIk5hbWUiOiAidGVzdCIsCiAgICAiR2xvYmFsUHJveHkiOiAidHJ1ZSIsCiAgICAiUmVtb3RlRG5zIjogIiIsCiAgICAiRG9tZXN0aWNEbnMiOiAiIiwKICAgICJHZW9pcHVybCI6ICIiLAogICAgIkdlb3NpdGV1cmwiOiAiIiwKICAgICJEbnNIb3N0cyI6IHt9LAogICAgIkRpcmVjdFNpdGVzIjogW10sCiAgICAiRGlyZWN0SXAiOiBbXSwKICAgICJQcm94eVNpdGVzIjogW10sCiAgICAiUHJveHlJcCI6IFtdLAogICAgIkJsb2NrU2l0ZXMiOiBbXSwKICAgICJCbG9ja0lwIjogW10sCiAgICAiRG9tYWluU3RyYXRlZ3kiOiAiQXNJcyIKfQ=="
)

// RoutingDeeplink builds happ://routing/add|onadd/{base64}.
func RoutingDeeplink(profileB64 string, autoActivate bool) string {
	profileB64 = strings.TrimSpace(profileB64)
	if profileB64 == "" {
		return DeeplinkScheme
	}
	if autoActivate {
		return RoutingOnAddPrefix + profileB64
	}
	return RoutingAddPrefix + profileB64
}

// OpenRoutingTarget is the happ:// target for redirect server.
func OpenRoutingTarget(profileB64 string, autoActivate bool) string {
	return RoutingDeeplink(profileB64, autoActivate)
}

// ResolveRoutingB64 returns user/env/default profile base64.
func ResolveRoutingB64(userB64, envDefault string) string {
	if s := strings.TrimSpace(userB64); s != "" {
		return s
	}
	if s := strings.TrimSpace(envDefault); s != "" {
		return s
	}
	return DefaultRoutingProfileB64
}

// ParseRoutingInput accepts happ://routing/… link, JSON profile, or raw base64.
func ParseRoutingInput(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("empty input")
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "happ://routing/") {
		if b64, ok := extractRoutingBase64FromLink(text); ok {
			return b64, nil
		}
		return "", fmt.Errorf("invalid happ routing link")
	}
	if strings.HasPrefix(text, "{") {
		if !json.Valid([]byte(text)) {
			return "", fmt.Errorf("invalid JSON profile")
		}
		return base64.StdEncoding.EncodeToString([]byte(text)), nil
	}
	if isLikelyBase64(text) {
		return strings.TrimSpace(text), nil
	}
	return "", fmt.Errorf("expected happ://routing/…, JSON, or base64 profile")
}

func extractRoutingBase64FromLink(link string) (string, bool) {
	lower := strings.ToLower(link)
	idx := strings.Index(lower, "happ://routing/")
	if idx < 0 {
		return "", false
	}
	rest := link[idx:]
	for _, p := range []string{"/onadd/", "/add/"} {
		pi := strings.Index(strings.ToLower(rest), p)
		if pi >= 0 {
			b64 := strings.TrimSpace(rest[pi+len(p):])
			if b64 != "" {
				return b64, true
			}
		}
	}
	return "", false
}

func isLikelyBase64(s string) bool {
	if len(s) < 16 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '+', r == '/', r == '=', r == '-', r == '_':
			continue
		default:
			return false
		}
	}
	return true
}
