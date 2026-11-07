package webshare

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"easysend/internal/item"
	"easysend/internal/transfer"
)

type stringErr string

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	sh, ok := s.validShare(w, r)
	if !ok {
		return
	}
	if !unlocked(r, sh.Token, sh.PIN) {
		http.Error(w, "PIN required", http.StatusUnauthorized)
		return
	}
	idx, err := strconv.Atoi(r.PathValue("idx"))
	if err != nil || idx < 0 || idx >= len(sh.Items) {
		http.NotFound(w, r)
		return
	}
	it := sh.Items[idx]
	inline := r.URL.Query().Get("inline") == "1"
	disp := "attachment"
	if inline {
		disp = "inline"
	}
	w.Header().Set("Content-Disposition", disp+"; filename*=UTF-8''"+url.PathEscape(it.BaseName()))
	w.Header().Set("Content-Type", it.MIME)
	if inline {
		// Previews don't count as downloads.
		serveItem(w, r, it, nil)
		return
	}
	id := s.cfg.Transfers.Start(transfer.Transfer{Direction: transfer.Outgoing, Method: transfer.Browser, Peer: clientName(r),
		Files: []transfer.File{{Name: it.Name, Size: it.Size}}, Status: transfer.Active}, nil)
	err = serveItem(w, r, it, func(n int64) { s.cfg.Transfers.AddBytes(id, 0, n) })
	s.finishDownload(sh, id, err, r)
}

func (s *Server) finishDownload(sh *Share, id string, err error, r *http.Request) {
	if err != nil || r.Context().Err() != nil {
		s.cfg.Transfers.Finish(id, transfer.Failed, errOr(err, "the browser stopped the download"))
		return
	}
	s.mu.Lock()
	sh.Downloads++
	s.mu.Unlock()
	s.cfg.Transfers.Finish(id, transfer.Done, nil)
}

func (e stringErr) Error() string { return string(e) }

func errOr(err error, msg string) error {
	if err != nil {
		return err
	}
	return stringErr(msg)
}

// serveItem sends an item with range support (resumable downloads,
// seeking in videos), counting the bytes sent.
func serveItem(w http.ResponseWriter, r *http.Request, it item.Item, count func(int64)) error {
	if it.Kind == item.Text {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		n, err := io.WriteString(w, it.Text)
		if count != nil {
			count(int64(n))
		}
		return err
	}
	f, err := os.Open(it.Path)
	if err != nil {
		http.Error(w, "file not available", http.StatusGone)
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	cw := &countingWriter{ResponseWriter: w, fn: count}
	http.ServeContent(cw, r, it.BaseName(), st.ModTime(), f)
	return nil
}
