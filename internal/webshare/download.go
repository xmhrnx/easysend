package webshare

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"easysend/internal/item"
	"easysend/internal/transfer"
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

type stringErr string

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

type countingWriter struct {
	http.ResponseWriter
	fn func(int64)
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.ResponseWriter.Write(p)
	if c.fn != nil && n > 0 {
		c.fn(int64(n))
	}
	return n, err
}

func (s *Server) handleZip(w http.ResponseWriter, r *http.Request) {
	sh, ok := s.validShare(w, r)
	if !ok {
		return
	}
	if !unlocked(r, sh.Token, sh.PIN) {
		http.Error(w, "PIN required", http.StatusUnauthorized)
		return
	}
	name := "EasySend-" + time.Now().Format("2006-01-02") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	files := make([]transfer.File, len(sh.Items))
	for i, it := range sh.Items {
		files[i] = transfer.File{Name: it.Name, Size: it.Size}
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	id := s.cfg.Transfers.Start(transfer.Transfer{Direction: transfer.Outgoing, Method: transfer.Browser, Peer: clientName(r), Files: files, Status: transfer.Active}, cancel)
	zw := zip.NewWriter(w)
	var err error
	for i, it := range sh.Items {
		if err = ctx.Err(); err != nil {
			break
		}
		if err = addToZip(zw, it, func(n int64) { s.cfg.Transfers.AddBytes(id, i, n) }); err != nil {
			break
		}
	}
	if err == nil {
		err = zw.Close()
	}
	s.finishDownload(sh, id, err, r)
}

func addToZip(zw *zip.Writer, it item.Item, count func(int64)) error {
	name := it.Name
	if it.Kind == item.Text {
		name = "Text.txt"
	}
	hdr := &zip.FileHeader{Name: name, Method: zip.Store, Modified: time.Now()}
	if strings.HasPrefix(it.MIME, "text/") {
		hdr.Method = zip.Deflate
	}
	dst, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	rc, err := it.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(&countingWriter{ResponseWriter: nopRW{dst}, fn: count}, rc)
	return err
}

type nopRW struct{ io.Writer }

func (nopRW) Header() http.Header { return http.Header{} }
func (nopRW) WriteHeader(int)     {}
