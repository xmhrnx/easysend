package webshare

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed pages/*.html
var pageFS embed.FS

var pages = template.Must(template.ParseFS(pageFS, "pages/*.html"))

type pageFile struct {
	Index int
	Name  string
	Base  string
	Dir   string
	Size  string
	Kind  string
	Text  string
}

type pageData struct {
	Device  string
	Token   string
	Files   []pageFile
	Total   string
	Count   int
	Expires string
	Error   string
	Back    string
}

func (s *Server) render(w http.ResponseWriter, name string, d pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := pages.ExecuteTemplate(w, name, d); err != nil {
		http.Error(w, "page error", http.StatusInternalServerError)
	}
}

func (s *Server) renderPIN(w http.ResponseWriter, back, msg string) {
	s.render(w, "pin", pageData{Device: s.cfg.Device(), Back: back, Error: msg})
}

func (s *Server) gone(w http.ResponseWriter) {
	w.WriteHeader(http.StatusGone)
	s.render(w, "gone", pageData{Device: s.cfg.Device()})
}

const logoSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#8B5CF6"/><stop offset=".55" stop-color="#6366F1"/><stop offset="1" stop-color="#22D3EE"/></linearGradient></defs><rect width="64" height="64" rx="16" fill="url(#g)"/><path d="M17 33.5 45.5 19.6c1.6-.8 3.2.8 2.4 2.4L34.1 50.5c-.8 1.6-3.1 1.4-3.6-.3l-3.2-10.4L17.4 36.6c-1.7-.5-1.9-2.7-.4-3.1Z" fill="#fff"/><path d="m27.3 39.8 9.4-9.4" stroke="#6366F1" stroke-width="2.6" stroke-linecap="round"/></svg>`
