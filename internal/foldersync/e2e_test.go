package foldersync

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"easysend/internal/localsend"
)

// device is a test EasySend: a nearby-protocol node with its own
// certificate, serving its folder sync manager.
type device struct {
	name   string
	fp     string
	node   *localsend.Node
	m      *Manager
	where  string // where it keeps invited folders
	accept bool
}

func (d *device) base() string { return "https://127.0.0.1:" + strconv.Itoa(d.node.Port()) }

func newDevice(t *testing.T, name string) *device {
	t.Helper()
	dir := t.TempDir()
	cert, fp, err := localsend.LoadOrCreateCert(dir)
	if err != nil {
		t.Fatal(err)
	}
	d := &device{name: name, fp: fp, where: filepath.Join(dir, "invited"), accept: true}
	var node *localsend.Node
	d.m, err = New(Config{
		Dir:  filepath.Join(dir, "sync"),
		Self: fp,
		Name: func() string { return name },
		Client: func(peer string) *http.Client {
			return node.PinnedClient(peer)
		},
		PeerFP: localsend.PeerFingerprint,
		OnInvite: func(_ context.Context, inv Invite) (string, bool) {
			return filepath.Join(d.where, inv.Name), d.accept
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	node, err = localsend.Start(localsend.Config{
		Self:        func() localsend.Info { return localsend.Info{Alias: name} },
		Cert:        &cert,
		Fingerprint: fp,
		Port:        -1,
		Routes:      d.m.Routes,
	})
	if err != nil {
		t.Fatal(err)
	}
	d.node = node
	t.Cleanup(func() { node.Close() })
	return d
}

// rounds syncs every device's folders a few times, as their loops would.
func rounds(devs ...*device) {
	for i := 0; i < 3; i++ {
		for _, d := range devs {
			d.m.round()
		}
	}
}

func write(t *testing.T, root, rel, s string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(root, rel string) string {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "<missing>"
	}
	return string(b)
}

func TestSyncEndToEnd(t *testing.T) {
	a, b := newDevice(t, "Alpha"), newDevice(t, "Beta")
	rootA := t.TempDir()
	write(t, rootA, "notes.txt", "hello")
	write(t, rootA, "photos/cat.jpg", "meow")

	f, err := a.m.Add("Shared", rootA)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.m.Invite(ctx, f.ID, b.fp, "Beta", b.base()); err != nil {
		t.Fatal(err)
	}
	// Beta knows Alpha by the address it last answered from.
	b.m.mu.Lock()
	b.m.folders[f.ID].Peers[0].Addr = a.base()
	b.m.mu.Unlock()
	rootB := filepath.Join(b.where, "Shared")

	rounds(a, b)
	if read(rootB, "notes.txt") != "hello" || read(rootB, "photos/cat.jpg") != "meow" {
		t.Fatalf("Beta didn't get the files: %q %q", read(rootB, "notes.txt"), read(rootB, "photos/cat.jpg"))
	}

	// Edits and new files on Beta reach Alpha.
	time.Sleep(10 * time.Millisecond) // a different modification time
	write(t, rootB, "notes.txt", "hello from beta")
	write(t, rootB, "todo.md", "- sync")
	rounds(b, a)
	if read(rootA, "notes.txt") != "hello from beta" || read(rootA, "todo.md") != "- sync" {
		t.Fatalf("Alpha didn't get Beta's changes: %q %q", read(rootA, "notes.txt"), read(rootA, "todo.md"))
	}

	// A deletion on Alpha reaches Beta, and the emptied folder goes.
	if err := os.Remove(filepath.Join(rootA, "photos", "cat.jpg")); err != nil {
		t.Fatal(err)
	}
	rounds(a, b)
	if _, err := os.Stat(filepath.Join(rootB, "photos")); !os.IsNotExist(err) {
		t.Fatalf("Beta kept the deleted file's folder: %v", err)
	}

	// Both edit the same file: the later edit wins on both, the other is
	// kept as a conflict copy on both.
	write(t, rootA, "notes.txt", "alpha's edit")
	time.Sleep(10 * time.Millisecond)
	write(t, rootB, "notes.txt", "beta's later edit")
	rounds(a, b)
	if read(rootA, "notes.txt") != "beta's later edit" || read(rootB, "notes.txt") != "beta's later edit" {
		t.Fatalf("conflict not resolved the same way: A %q, B %q", read(rootA, "notes.txt"), read(rootB, "notes.txt"))
	}
	for _, root := range []string{rootA, rootB} {
		copies, _ := filepath.Glob(filepath.Join(root, "notes (conflict, Alpha *).txt"))
		if len(copies) != 1 || read(root, filepath.Base(copies[0])) != "alpha's edit" {
			t.Fatalf("conflict copy missing in %s: %v", root, copies)
		}
	}

	// Nothing left to do: another round fetches nothing.
	if st := a.m.Folders()[0].Status; st.Err != "" || st.Peers[b.fp].Err != "" {
		t.Fatalf("errors after syncing: %+v", st)
	}
}

func TestStrangersAreTurnedAway(t *testing.T) {
	a, b, c := newDevice(t, "Alpha"), newDevice(t, "Beta"), newDevice(t, "Carol")
	f, err := a.m.Add("Private", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.m.Invite(ctx, f.ID, b.fp, "Beta", b.base()); err != nil {
		t.Fatal(err)
	}
	// Carol isn't a member: Alpha won't hand her the index.
	_, err = c.m.fetchIndex(ctx, a.fp, a.base(), f.ID)
	var se *statusErr
	if !errors.As(err, &se) || se.code != 403 {
		t.Fatalf("a stranger read the folder: %v", err)
	}
	// Pinned: Beta's client won't accept Carol's server posing as Alpha.
	_, err = b.m.fetchIndex(ctx, a.fp, c.base(), f.ID)
	if err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("talked to the wrong device: %v", err)
	}
	// A declined invitation adds nobody.
	c.accept = false
	if err := a.m.Invite(ctx, f.ID, c.fp, "Carol", c.base()); !errors.Is(err, ErrDeclined) {
		t.Fatalf("decline: %v", err)
	}
	if peers := a.m.Folders()[0].Peers; len(peers) != 1 {
		t.Fatalf("declined device added: %+v", peers)
	}
}
