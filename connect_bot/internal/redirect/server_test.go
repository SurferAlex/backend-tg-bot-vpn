package redirect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connect-bot/internal/happ"
)

func TestHandleOpen_redirectsRoutingOnAdd(t *testing.T) {
	s := NewServer("", true, "")
	req := httptest.NewRequest(http.MethodGet, "/open?routing=abc123", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	want := happ.RoutingOnAddPrefix + "abc123"
	if loc != want {
		t.Fatalf("Location %q want %q", loc, want)
	}
}

func TestHandleOpen_htmlFallback(t *testing.T) {
	s := NewServer("", true, "")
	req := httptest.NewRequest(http.MethodGet, "/open?routing=abc&html=1", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), happ.RoutingOnAddPrefix) {
		t.Fatal("expected routing deeplink in HTML")
	}
}
