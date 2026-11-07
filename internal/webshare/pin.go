package webshare

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
)

func cookieName(token string) string { return "es_" + token[:8] }

func pinCookie(token, pin string) string {
	sum := sha256.Sum256([]byte(token + "\x00" + pin))
	return hex.EncodeToString(sum[:])
}

// unlocked reports whether the request may use token (PIN given or none
// needed).
func unlocked(r *http.Request, token, pin string) bool {
	if pin == "" {
		return true
	}
	c, err := r.Cookie(cookieName(token))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(pinCookie(token, pin))) == 1
}

var errLocked = errors.New("too many wrong PINs")

func (s *Server) handlePIN(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	var want, back string
	if sh, ok := s.share(token); ok {
		want, back = sh.PIN, "/s/"+token
	} else if token == s.cfg.DropToken() {
		want, back = s.cfg.DropPIN(), "/r/"+token
	} else {
		s.gone(w)
		return
	}
	s.mu.Lock()
	locked := s.fails[token] >= 10
	s.mu.Unlock()
	if locked {
		s.renderPIN(w, back, errLocked.Error())
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.FormValue("pin")), []byte(want)) != 1 {
		s.mu.Lock()
		s.fails[token]++
		s.mu.Unlock()
		s.renderPIN(w, back, "That PIN isn't right. Try again.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName(token), Value: pinCookie(token, want), Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, back, http.StatusSeeOther)
}

func clientName(r *http.Request) string {
	ua := r.UserAgent()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return browserName(ua) + " · " + host
}
