package happ

import (
	"strings"
	"testing"
)

func TestInlineOpenURL_vlessParam(t *testing.T) {
	got := InlineOpenURL(DeliveryOptions{
		OpenRedirectBase: "https://example.com/happ/open",
		VlessURL:         "vless://u@h:443",
	})
	if !strings.Contains(got, "vless=vless%3A%2F%2F") {
		t.Fatalf("got %q", got)
	}
}

func TestInlineOpenURL_routingWhenNoVless(t *testing.T) {
	got := InlineOpenURL(DeliveryOptions{
		OpenRedirectBase: "https://example.com/happ/open",
		RoutingB64:       "abc",
	})
	if !strings.Contains(got, "routing=abc") {
		t.Fatalf("got %q", got)
	}
}
