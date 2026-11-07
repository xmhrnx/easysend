package webshare

import (
	"net/http"
)

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
