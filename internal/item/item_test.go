package item

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoin(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"../../etc/passwd":  "etc/passwd",
		"/abs/file.txt":     "abs/file.txt",
		`C:\Windows\a.txt`:  "C_/Windows/a.txt",
		"Photos/2024/a.jpg": "Photos/2024/a.jpg",
		"..":                "file",
		"a/./b/../c.txt":    "a/c.txt",
	}
	for in, want := range cases {
		got := SafeJoin(dir, in)
		rel, _ := filepath.Rel(dir, got)
		if filepath.ToSlash(rel) != want {
			t.Errorf("SafeJoin(%q) = %q, want %q", in, rel, want)
		}
		if !strings.HasPrefix(got, dir) {
			t.Errorf("escaped: %q", got)
		}
	}
}

func TestUniqueTopLevel(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "Photos"), 0o755); err != nil {
		t.Fatal(err)
	}
	ren := map[string]string{}
	a := UniqueTopLevel(dir, "Photos/a.jpg", ren)
	b := UniqueTopLevel(dir, "Photos/sub/b.jpg", ren)
	if filepath.Base(filepath.Dir(a)) != "Photos (2)" || !strings.Contains(b, "Photos (2)") {
		t.Fatalf("got %q %q", a, b)
	}
}

func TestFromFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Top")
	_ = os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "sub", "x.txt"), []byte("hi"), 0o644)
	items, err := FromFolder(dir)
	if err != nil || len(items) != 1 || items[0].Name != "Top/sub/x.txt" || items[0].Size != 2 {
		t.Fatalf("%v %+v", err, items)
	}
}

func TestHumanSize(t *testing.T) {
	if HumanSize(1536) != "1.5 KB" || HumanSize(10) != "10 B" {
		t.Fatal(HumanSize(1536))
	}
}
