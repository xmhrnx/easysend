package webshare

import (
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"easysend/internal/item"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.render(w, "home", pageData{Device: s.cfg.Device()})
}

func (s *Server) validShare(w http.ResponseWriter, r *http.Request) (*Share, bool) {
	sh, ok := s.share(r.PathValue("token"))
	if !ok || sh.Expired() {
		s.gone(w)
		return nil, false
	}
	return sh, true
}

func (s *Server) handleShare(w http.ResponseWriter, r *http.Request) {
	sh, ok := s.validShare(w, r)
	if !ok {
		return
	}
	if !unlocked(r, sh.Token, sh.PIN) {
		s.renderPIN(w, "/s/"+sh.Token, "")
		return
	}
	s.mu.Lock()
	snap := *sh
	s.mu.Unlock()
	d := pageData{Device: s.cfg.Device(), Token: snap.Token, Total: item.HumanSize(snap.Size()), Count: len(snap.Items)}
	for i, it := range snap.Items {
		f := pageFile{Index: i, Name: it.Name, Base: it.BaseName(), Size: item.HumanSize(it.Size), Kind: kindOf(it), Text: it.Text}
		if d := path.Dir(it.Name); d != "." {
			f.Dir = d
		}
		d.Files = append(d.Files, f)
	}
	if !snap.Expires.IsZero() {
		d.Expires = "Expires " + humanUntil(snap.Expires)
	}
	s.render(w, "share", d)
}

func kindOf(it item.Item) string {
	switch {
	case it.Kind == item.Text:
		return "text"
	case strings.HasPrefix(it.MIME, "image/"):
		return "image"
	case strings.HasPrefix(it.MIME, "video/"):
		return "video"
	case strings.HasPrefix(it.MIME, "audio/"):
		return "audio"
	}
	return "file"
}

func humanUntil(t time.Time) string {
	d := time.Until(t)
	switch {
	case d < time.Minute:
		return "in under a minute"
	case d < time.Hour:
		return "in " + strconv.Itoa(int(d.Minutes())) + " min"
	case d < 48*time.Hour:
		return "in " + strconv.Itoa(int(d.Hours()+0.5)) + " h"
	}
	return "on " + t.Format("Jan 2")
}
