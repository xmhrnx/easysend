package webshare

import (
	"net/http"
	"strings"

	"easysend/internal/diag"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /s/{token}", s.handleShare)
	mux.HandleFunc("POST /s/{token}/pin", s.handlePIN)
	mux.HandleFunc("GET /s/{token}/f/{idx}", s.handleFile)
	mux.HandleFunc("GET /s/{token}/zip", s.handleZip)
	mux.HandleFunc("GET /r/{token}", s.handleDrop)
	mux.HandleFunc("POST /r/{token}/pin", s.handlePIN)
	mux.HandleFunc("POST /r/{token}/begin", s.handleBegin)
	mux.HandleFunc("POST /r/{token}/file", s.handleUpload)
	mux.HandleFunc("POST /r/{token}/text", s.handleText)
	mux.HandleFunc("GET /logo.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "max-age=86400")
		_, _ = w.Write([]byte(logoSVG))
	})
	return securityHeaders(mux)
}

func securityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		diag.Logf("browser: %s %s from %s", r.Method, redact(r.URL.Path), r.RemoteAddr)
		h.ServeHTTP(w, r)
	})
}

// redact keeps a request path's kind (/s/, /r/) without its secret token.
func redact(path string) string {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 3)
	if len(parts) >= 2 && (parts[0] == "s" || parts[0] == "r") {
		return "/" + parts[0] + "/…"
	}
	return path
}
