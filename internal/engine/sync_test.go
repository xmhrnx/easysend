package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"easysend/internal/settings"
)

func TestFolderSync(t *testing.T) {
	a, b := newEngine(t, "Alice"), newEngine(t, "Bob")
	// Auto-accepting files never accepts a folder.
	_ = b.UpdateSettings(func(s *settings.Settings) { s.AutoAccept = settings.AcceptEveryone })
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	dev, err := a.AddByAddress(ctx, "127.0.0.1:"+strconv.Itoa(b.Self().Port))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.AddByAddress(ctx, "127.0.0.1:"+strconv.Itoa(a.Self().Port)); err != nil {
		t.Fatal(err)
	}
	rootA := filepath.Join(t.TempDir(), "Projects")
	if err := os.MkdirAll(rootA, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(rootA, "plan.txt"), []byte("v1"), 0o644)
	f, err := a.AddSyncFolder("Projects", rootA)
	if err != nil {
		t.Fatal(err)
	}

	invited := make(chan error, 1)
	go func() { invited <- a.InviteToSync(ctx, f.ID, Peer{Device: dev}) }()
	waitFor(t, func() bool { return len(b.Requests()) == 1 })
	req := b.Requests()[0]
	if req.SyncFolder != "Projects" || req.Sender.Alias != "Alice" {
		t.Fatalf("%+v", req)
	}
	rootB := filepath.Join(t.TempDir(), "Elsewhere")
	b.AnswerSync(req.ID, true, rootB)
	if err := <-invited; err != nil {
		t.Fatal(err)
	}

	read := func(p string) string { s, _ := os.ReadFile(p); return string(s) }
	waitFor(t, func() bool { return read(filepath.Join(rootB, "plan.txt")) == "v1" })
	_ = os.WriteFile(filepath.Join(rootB, "notes.txt"), []byte("from bob"), 0o644)
	b.SyncNow()
	waitFor(t, func() bool { return read(filepath.Join(rootA, "notes.txt")) == "from bob" })

	// A declined invitation.
	c := newEngine(t, "Carol")
	devC, err := a.AddByAddress(ctx, "127.0.0.1:"+strconv.Itoa(c.Self().Port))
	if err != nil {
		t.Fatal(err)
	}
	go func() { invited <- a.InviteToSync(ctx, f.ID, Peer{Device: devC}) }()
	waitFor(t, func() bool { return len(c.Requests()) == 1 })
	c.AnswerSync(c.Requests()[0].ID, false, "")
	if err := <-invited; !errors.Is(err, ErrSyncDeclined) {
		t.Fatal(err)
	}
}
