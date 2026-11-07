package localsend

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
)

func (n *Node) routes() http.Handler {
	mux := http.NewServeMux()
	for _, v := range []string{"/api/localsend/v1", apiBase} {
		mux.HandleFunc(v+"/info", n.handleInfo)
		mux.HandleFunc(v+"/register", n.handleRegister)
	}
	mux.HandleFunc("POST "+apiBase+"/prepare-upload", n.handlePrepare)
	mux.HandleFunc("POST "+apiBase+"/upload", n.handleUpload)
	mux.HandleFunc("POST "+apiBase+"/cancel", n.handleCancel)
	return mux
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (n *Node) discoverable() bool {
	return n.cfg.Discoverable == nil || n.cfg.Discoverable()
}

func (n *Node) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("fingerprint") == n.cfg.Fingerprint {
		http.Error(w, "self-discovered", http.StatusPreconditionFailed)
		return
	}
	if !n.discoverable() {
		http.Error(w, "hidden", http.StatusForbidden)
		return
	}
	writeJSON(w, n.Self())
}

func (n *Node) handleRegister(w http.ResponseWriter, r *http.Request) {
	var info Info
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&info); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if info.Fingerprint == n.cfg.Fingerprint {
		http.Error(w, "self-discovered", http.StatusPreconditionFailed)
		return
	}
	n.addDevice(Device{Info: info, IP: remoteIP(r)})
	if !n.discoverable() {
		http.Error(w, "hidden", http.StatusForbidden)
		return
	}
	writeJSON(w, n.Self())
}

func newToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
