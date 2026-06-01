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
	addr              string
	mux               *http.ServeMux
	routingAutoOnAdd  bool
	defaultRoutingB64 string
}

func NewServer(listenAddr string, routingAutoOnAdd bool, defaultRoutingB64 string) *Server {
	addr := strings.TrimSpace(listenAddr)
	if addr == "" {
		addr = defaultListen
	}
	s := &Server{
		addr:              addr,
		mux:               http.NewServeMux(),
		routingAutoOnAdd:  routingAutoOnAdd,
		defaultRoutingB64: strings.TrimSpace(defaultRoutingB64),
	}
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

	vless := r.URL.Query().Get("vless")
	routing := r.URL.Query().Get("routing")
	target, _ := happ.ResolveOpenTarget(vless, routing, s.routingAutoOnAdd, s.defaultRoutingB64)

	// ?redirect=1 — чистый 302 (для curl); по умолчанию HTML+JS для Telegram/WKWebView.
	useHTML := r.URL.Query().Get("redirect") != "1"
	if r.Method == http.MethodGet && (useHTML || r.URL.Query().Get("html") == "1") {
		writeOpenHTML(w, target)
		return
	}

	if r.Method == http.MethodHead {
		w.Header().Set("Location", target)
		w.WriteHeader(http.StatusFound)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func writeOpenHTML(w http.ResponseWriter, target string) {
	escaped := htmlEscape(target)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Открыть Happ</title>
</head>
<body>
<p>Открываем Happ…</p>
<p><a id="open" href="%s" style="font-size:18px">Открыть Happ</a></p>
<script>
(function () {
  var u = %q;
  try { window.location.replace(u); } catch (e) {}
  setTimeout(function () {
    var a = document.getElementById("open");
    if (a) a.focus();
  }, 300);
})();
</script>
</body>
</html>`, escaped, target)
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
