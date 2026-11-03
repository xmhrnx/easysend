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

const logoSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><defs><linearGradient id="b" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2A231A"/><stop offset="1" stop-color="#0B0907"/></linearGradient><linearGradient id="g" x1=".2" y1="0" x2=".8" y2="1"><stop offset="0" stop-color="#FFF1C4"/><stop offset=".45" stop-color="#F0C964"/><stop offset="1" stop-color="#B07C26"/></linearGradient></defs><rect width="64" height="64" rx="16" fill="url(#b)"/><rect x=".5" y=".5" width="63" height="63" rx="15.5" fill="none" stroke="#F2D27C" stroke-opacity=".35"/><path d="M14 31 47.5 15.9c1.7-.8 3.4.8 2.6 2.5L35.3 51.6c-.8 1.7-3.2 1.6-3.7-.2l-3.5-12.6Z" fill="url(#g)"/><path d="m28.1 38.8 22-20.4" stroke="#8C6A2A" stroke-width="1.6" stroke-linecap="round"/></svg>`
