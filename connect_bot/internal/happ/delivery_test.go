package happ

import (
	"strings"
	"testing"
)

func TestBuildMessageText_routing(t *testing.T) {
	text := BuildMessageText(DeliveryOptions{})
	if !strings.Contains(text, "routing") {
		t.Fatal("expected routing mention")
	}
}

func TestInlineDownloadURL_isAppStore(t *testing.T) {
	got := InlineDownloadURL(DeliveryOptions{})
	if !strings.Contains(got, "apps.apple.com") {
		t.Fatalf("got %q", got)
	}
}

func TestInlineOpenURL_redirectRouting(t *testing.T) {
	got := InlineOpenURL(DeliveryOptions{
		OpenRedirectBase: "https://example.com/happ/open",
		RoutingB64:       "abc123",
	})
	if !strings.Contains(got, "routing=abc123") {
		t.Fatalf("got %q", got)
	}
}

func TestInlineOpenURL_defaultProfile(t *testing.T) {
	got := InlineOpenURL(DeliveryOptions{OpenRedirectBase: "https://example.com/happ/open"})
	if !strings.Contains(got, "routing=") {
		t.Fatalf("got %q", got)
	}
}
