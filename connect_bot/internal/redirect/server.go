package redirect

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"connect-bot/internal/happ"
)

const defaultListen = ":8091"

// Server serves HTTPS-facing redirect to happ:// (opened from Telegram inline buttons).
type Server struct {
	addr string
	mux  *http.ServeMux
}

func NewServer(listenAddr string) *Server {
	addr := strings.TrimSpace(listenAddr)
	if addr == "" {
		addr = defaultListen
	}
	s := &Server{addr: addr, mux: http.NewServeMux()}
	s.mux.HandleFunc("/open", s.handleOpen)
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return s
}

func (s *Server) handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	config := r.URL.Query().Get("vless")
	if config == "" {
		config = r.URL.Query().Get("sub")
	}
	target := happ.OpenAppURL(config)
	if r.Method == http.MethodHead {
		w.Header().Set("Location", target)
		w.WriteHeader(http.StatusFound)
		return
	}
	writeOpenHTML(w, target)
}

func writeOpenHTML(w http.ResponseWriter, target string) {
	escaped := htmlEscape(target)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	// Some in-app browsers follow 302 to custom schemes; HTML fallback helps Safari.
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="0;url=%s">
<title>Открыть Happ</title>
</head>
<body>
<p>Открываем Happ…</p>
<p><a href="%s">Нажмите здесь, если приложение не открылось</a></p>
<p>Подтвердите добавление конфигурации в Happ.</p>
</body>
</html>`, escaped, escaped)
}

func htmlEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '"':
			b.WriteString("&quot;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Run listens until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("happ redirect: listening on %s (GET /open)", s.addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
