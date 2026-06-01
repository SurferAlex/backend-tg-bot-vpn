package redirect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleOpen_containsHappAddVless(t *testing.T) {
	s := NewServer("")
	req := httptest.NewRequest(http.MethodGet, "/open?vless=vless%3A%2F%2Ftest%40host%3A443", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "happ://add/vless://") {
		t.Fatalf("expected happ://add/vless:// in HTML, got %q", body)
	}
}
