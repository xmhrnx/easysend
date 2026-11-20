package foldersync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func file(hash string, changed int64, by string) Entry {
	return Entry{Hash: hash, Size: 1, Changed: changed, By: by}
}

func gone(changed int64, by string) Entry { return Entry{Deleted: true, Changed: changed, By: by} }

func TestPlan(t *testing.T) {
	cases := []struct {
		name          string
		local, remote Index
		base          map[string]string
		want          actionKind
		none          bool
	}{
		{name: "same version, base behind", local: Index{"a": file("h1", 1, "A")}, remote: Index{"a": file("h1", 2, "B")}, want: agree},
		{name: "same version, base current", local: Index{"a": file("h1", 1, "A")}, remote: Index{"a": file("h1", 2, "B")}, base: map[string]string{"a": "h1"}, none: true},
		{name: "new on the peer", local: Index{}, remote: Index{"a": file("h1", 1, "B")}, want: fetch},
		{name: "deleted on the peer, never here", local: Index{}, remote: Index{"a": gone(1, "B")}, want: agree},
		{name: "only the peer edited", local: Index{"a": file("h1", 1, "A")}, remote: Index{"a": file("h2", 2, "B")}, base: map[string]string{"a": "h1"}, want: fetch},
		{name: "only the peer deleted", local: Index{"a": file("h1", 1, "A")}, remote: Index{"a": gone(2, "B")}, base: map[string]string{"a": "h1"}, want: remove},
		{name: "only this device edited", local: Index{"a": file("h2", 2, "A")}, remote: Index{"a": file("h1", 1, "B")}, base: map[string]string{"a": "h1"}, none: true},
		{name: "edited here, deleted there", local: Index{"a": file("h2", 3, "A")}, remote: Index{"a": gone(2, "B")}, base: map[string]string{"a": "h1"}, none: true},
		{name: "deleted here, edited there", local: Index{"a": gone(2, "A")}, remote: Index{"a": file("h2", 3, "B")}, base: map[string]string{"a": "h1"}, want: fetch},
		{name: "both edited, peer later", local: Index{"a": file("h2", 2, "A")}, remote: Index{"a": file("h3", 3, "B")}, base: map[string]string{"a": "h1"}, want: keepBoth},
		{name: "both edited, this later", local: Index{"a": file("h2", 4, "A")}, remote: Index{"a": file("h3", 3, "B")}, base: map[string]string{"a": "h1"}, none: true},
		{name: "no common base, peer later", local: Index{"a": file("h2", 2, "A")}, remote: Index{"a": file("h3", 3, "B")}, want: keepBoth},
		{name: "only here", local: Index{"a": file("h2", 2, "A")}, remote: Index{}, none: true},
	}
	for _, c := range cases {
		acts := plan(c.local, c.remote, c.base)
		if c.none {
			if len(acts) != 0 {
				t.Errorf("%s: want no action, got %+v", c.name, acts)
			}
			continue
		}
		if len(acts) != 1 || acts[0].kind != c.want {
			t.Errorf("%s: want %v, got %+v", c.name, c.want, acts)
		}
	}
}

// TestPlanSymmetric checks both devices resolve a double edit the same
// way: exactly one keeps a conflict copy.
func TestPlanSymmetric(t *testing.T) {
	a, b := file("ha", 5, "A"), file("hb", 5, "B") // same time: the device ID decides
	base := map[string]string{"x": "h0"}
	onA := plan(Index{"x": a}, Index{"x": b}, base)
	onB := plan(Index{"x": b}, Index{"x": a}, base)
	if (len(onA) == 1) == (len(onB) == 1) {
		t.Fatalf("exactly one side should yield: A %+v, B %+v", onA, onB)
	}
}

func TestScan(t *testing.T) {
	root := t.TempDir()
	write := func(rel, s string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt", "one")
	write("dir/b.txt", "two")
	write(".DS_Store", "junk")
	write(internalDir+"/tmp", "temp")
	now := time.Unix(100, 0)
	idx, changed, err := scan(root, nil, "A", now)
	if err != nil || !changed || len(idx) != 2 || idx["dir/b.txt"].Hash == "" {
		t.Fatalf("first scan: %v %v %+v", err, changed, idx)
	}
	again, changed, _ := scan(root, idx, "A", now.Add(time.Minute))
	if changed || again["a.txt"] != idx["a.txt"] {
		t.Fatalf("unchanged folder rescanned as changed: %+v", again)
	}
	_ = os.Remove(filepath.Join(root, "a.txt"))
	write("dir/b.txt", "two, edited")
	later := now.Add(2 * time.Minute)
	third, changed, _ := scan(root, again, "A", later)
	if !changed || !third["a.txt"].Deleted || third["a.txt"].Changed != later.UnixNano() {
		t.Fatalf("deletion not recorded: %+v", third["a.txt"])
	}
	if third["dir/b.txt"].Hash == idx["dir/b.txt"].Hash || third["dir/b.txt"].Changed != later.UnixNano() {
		t.Fatalf("edit not recorded: %+v", third["dir/b.txt"])
	}
}

func TestCleanPath(t *testing.T) {
	for p, ok := range map[string]bool{
		"a.txt": true, "dir/b.txt": true,
		"": false, "/etc/passwd": false, "../x": false, "a/../../x": false, "a//b": false,
		`a\b`: false, "C:/x": false, internalDir + "/t": false, "./a": false,
	} {
		if (cleanPath(p) != "") != ok {
			t.Errorf("cleanPath(%q) = %q", p, cleanPath(p))
		}
	}
}
