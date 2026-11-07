package webshare

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"easysend/internal/item"
	"easysend/internal/transfer"
)

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
