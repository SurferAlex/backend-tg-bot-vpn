package happ

import (
	"strings"
	"testing"
)

func TestBuildMessageText_clipboardReady(t *testing.T) {
	text := BuildMessageText(DeliveryOptions{})
	if strings.Contains(text, "vless://942b") {
		t.Fatal("must not embed user key in message text")
	}
	if !strings.Contains(text, "Скачать Happ") {
		t.Fatal("expected download step")
	}
}

func TestInlineDownloadURL_isAppStore(t *testing.T) {
	got := InlineDownloadURL(DeliveryOptions{})
	if !strings.Contains(got, "apps.apple.com") {
		t.Fatalf("download must be App Store, got %q", got)
	}
}

func TestInlineOpenURL_redirectWithUserVless(t *testing.T) {
	vless := "vless://uuid@host:443?type=tcp"
	got := InlineOpenURL(DeliveryOptions{
		OpenRedirectBase: "https://example.com/happ/open",
		VlessURL:         vless,
	})
	if !strings.HasPrefix(got, "https://example.com/happ/open?vless=") {
		t.Fatalf("got %q", got)
	}
}

func TestInlineOpenURL_defaultRedirectBase(t *testing.T) {
	got := InlineOpenURL(DeliveryOptions{VlessURL: "vless://a@b:443"})
	if !strings.HasPrefix(got, DefaultOpenRedirectPublicURL+"?vless=") {
		t.Fatalf("got %q", got)
	}
}

func TestInlineOpenURL_noVlessAppStore(t *testing.T) {
	got := InlineOpenURL(DeliveryOptions{})
	if !strings.Contains(got, "apps.apple.com") {
		t.Fatalf("without vless must be App Store, got %q", got)
	}
}

func TestAddConfigURL_vless(t *testing.T) {
	vless := "vless://942b9878-5017-473a-9285-a9391558f267@panel.example:443?type=tcp"
	got := AddConfigURL(vless)
	want := DeeplinkAddPrefix + vless
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestOpenRedirectURL_encodesVless(t *testing.T) {
	got := OpenRedirectURL("https://host/happ/open", "vless://a@b:443?x=1")
	if !strings.Contains(got, "vless%3A%2F%2F") {
		t.Fatalf("expected encoded vless, got %q", got)
	}
}
