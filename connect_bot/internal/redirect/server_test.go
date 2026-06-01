package redirect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connect-bot/internal/happ"
)

func TestHandleOpen_vlessHTML(t *testing.T) {
	s := NewServer("", true, "")
	req := httptest.NewRequest(http.MethodGet, "/open?vless=vless%3A%2F%2Fu%40h%3A443", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "happ://add/vless://u@h:443") {
		t.Fatalf("body missing deeplink: %s", body)
	}
	if !strings.Contains(body, "window.location.replace") {
		t.Fatal("expected JS redirect")
	}
}

func TestHandleOpen_vless302(t *testing.T) {
	s := NewServer("", true, "")
	req := httptest.NewRequest(http.MethodGet, "/open?vless=vless%3A%2F%2Fu%40h%3A443&redirect=1", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status %d", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "happ://add/vless://u@h:443" {
		t.Fatalf("Location %q", got)
	}
}

func TestHandleOpen_routingIgnoredWhenVlessEmptyUsesDefault(t *testing.T) {
	s := NewServer("", true, "")
	req := httptest.NewRequest(http.MethodGet, "/open", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), happ.RoutingOnAddPrefix) {
		t.Fatal("expected default routing deeplink")
	}
}
