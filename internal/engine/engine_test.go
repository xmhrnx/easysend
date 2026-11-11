package engine

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"easysend/internal/item"
	"easysend/internal/localsend"
	"easysend/internal/settings"
	"easysend/internal/transfer"
)

func newEngine(t *testing.T, name string) *Engine {
	t.Helper()
	dir := t.TempDir()
	// Settings with UPnP off and free ports, before the engine starts.
	st, err := settings.Open(dir, filepath.Join(dir, "dl"))
	if err != nil {
		t.Fatal(err)
	}
	_ = st.Update(func(s *settings.Settings) { s.UPnP = false; s.Port = -1; s.WebPort = 0; s.DeviceName = name })
	e, err := Start(Config{DataDir: dir, DownloadDir: filepath.Join(dir, "dl"), DeviceModel: "Test", DeviceType: localsend.TypeDesktop})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	return e
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if cond() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("timed out")
}

func TestNearbyRequestFlow(t *testing.T) {
	a, b := newEngine(t, "Alice"), newEngine(t, "Bob")
	ctx := context.Background()
	dev, err := a.AddByAddress(ctx, "127.0.0.1:"+strconv.Itoa(b.Self().Port))
	if err != nil || dev.Alias != "Bob" {
		t.Fatal(err, dev)
	}
	src := filepath.Join(t.TempDir(), "photo.jpg")
	_ = os.WriteFile(src, []byte("jpegdata"), 0o644)
	it, _ := item.FromFile(src)
	id := a.SendNearby(dev, []item.Item{it}, "")

	waitFor(t, func() bool { return len(b.Requests()) == 1 })
	req := b.Requests()[0]
	if req.Sender.Alias != "Alice" || req.Total != 8 {
		t.Fatalf("%+v", req)
	}
	b.Answer(req.ID, true)
	waitFor(t, func() bool { tr, _ := a.Transfers.Get(id); return tr.Status.Finished() })
	if tr, _ := a.Transfers.Get(id); tr.Status != transfer.Done {
		t.Fatalf("%+v", tr)
	}
	got, err := os.ReadFile(filepath.Join(b.Settings.Get().DownloadDir, "photo.jpg"))
	if err != nil || string(got) != "jpegdata" {
		t.Fatal(err)
	}

	// Decline, then PIN.
	id = a.SendNearby(dev, []item.Item{it}, "")
	waitFor(t, func() bool { return len(b.Requests()) == 1 })
	b.Answer(b.Requests()[0].ID, false)
	waitFor(t, func() bool { tr, _ := a.Transfers.Get(id); return tr.Status == transfer.Declined })

	_ = b.UpdateSettings(func(s *settings.Settings) {
		s.RequirePIN = true
		s.PIN = "9999"
		s.AutoAccept = settings.AcceptEveryone
	})
	id = a.SendNearby(dev, []item.Item{it}, "")
	waitFor(t, func() bool { tr, _ := a.Transfers.Get(id); return IsPINError(tr) })
	id = a.SendNearby(dev, []item.Item{item.FromText("hey")}, "9999")
	waitFor(t, func() bool { tr, _ := a.Transfers.Get(id); return tr.Status == transfer.Done })
	waitFor(t, func() bool {
		for _, tr := range b.Transfers.List() {
			if tr.Text == "hey" && tr.Peer == "Alice" {
				return true
			}
		}
		return false
	})
}

func TestShareLinks(t *testing.T) {
	e := newEngine(t, "Carol")
	v := e.CreateShare([]item.Item{item.FromText("x")}, ShareOptions{Expires: time.Hour})
	if v.LANURL == "" || v.PublicURL != "" || len(e.Shares()) != 1 {
		t.Fatalf("%+v", v)
	}
	_ = e.UpdateSettings(func(s *settings.Settings) { s.PublicURL = "https://files.example.com/" })
	if got := e.Shares()[0].PublicURL; got != "https://files.example.com/s/"+v.Token {
		t.Fatal(got)
	}
	e.StopShare(v.Token)
	if len(e.Shares()) != 0 {
		t.Fatal("not stopped")
	}
}

func TestPublishMovesReceivedFiles(t *testing.T) {
	a, b := newEngine(t, "Alice"), newEngine(t, "Bob")
	published := t.TempDir()
	b.cfg.Publish = func(ctx context.Context, p string) (string, error) {
		dst := filepath.Join(published, filepath.Base(p))
		data, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		return "content://downloads/" + filepath.Base(p), os.WriteFile(dst, data, 0o644)
	}
	_ = b.UpdateSettings(func(s *settings.Settings) { s.AutoAccept = settings.AcceptEveryone })
	dev, err := a.AddByAddress(context.Background(), "127.0.0.1:"+strconv.Itoa(b.Self().Port))
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "doc.pdf")
	_ = os.WriteFile(src, []byte("pdf"), 0o644)
	it, _ := item.FromFile(src)
	a.SendNearby(dev, []item.Item{it}, "")
	waitFor(t, func() bool {
		for _, tr := range b.Transfers.List() {
			if tr.Direction == transfer.Incoming && len(tr.Files) == 1 && tr.Files[0].Saved == "content://downloads/doc.pdf" {
				return true
			}
		}
		return false
	})
	if _, err := os.Stat(filepath.Join(b.Settings.Get().DownloadDir, "doc.pdf")); !os.IsNotExist(err) {
		t.Fatal("original not removed")
	}
	if got, _ := os.ReadFile(filepath.Join(published, "doc.pdf")); string(got) != "pdf" {
		t.Fatal("not published")
	}
}

// TestFixedDownloadDir checks a saved folder from an earlier install (an
// old iOS container) gives way to the current one.
func TestFixedDownloadDir(t *testing.T) {
	dir := t.TempDir()
	st, err := settings.Open(dir, "/old/container/Documents/EasySend")
	if err != nil {
		t.Fatal(err)
	}
	_ = st.Update(func(s *settings.Settings) { s.UPnP = false; s.Port = -1; s.WebPort = 0 })
	now := filepath.Join(dir, "Documents", "EasySend")
	e, err := Start(Config{DataDir: dir, DownloadDir: now, FixedDownloadDir: true, DeviceType: localsend.TypeMobile})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if got := e.Settings.Get().DownloadDir; got != now {
		t.Fatalf("download dir %q, want %q", got, now)
	}
}
