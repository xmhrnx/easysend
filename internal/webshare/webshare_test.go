package webshare

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"easysend/internal/item"
	"easysend/internal/transfer"
)

func start(t *testing.T) (*Server, string, string) {
	dl := t.TempDir()
	reg := transfer.NewRegistry(t.TempDir(), nil)
	s, err := Start(Config{
		Port: 0, Device: func() string { return "Test" }, Transfers: reg,
		SaveDir:     func() string { return dl },
		DropEnabled: func() bool { return true }, DropPIN: func() string { return "" },
		DropToken: func() string { return "droptok" },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, "http://127.0.0.1:" + strconv.Itoa(s.Port()), dl
}

func get(t *testing.T, c *http.Client, u string) (int, string) {
	resp, err := c.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestShareDownloadWithPIN(t *testing.T) {
	s, base, _ := start(t)
	p := filepath.Join(t.TempDir(), "hello.txt")
	_ = os.WriteFile(p, []byte("hello world"), 0o644)
	it, _ := item.FromFile(p)
	sh := s.Add(Share{Items: []item.Item{it, item.FromText("note")}, PIN: "1234"})

	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar}
	if code, body := get(t, c, base+"/s/"+sh.Token); code != 200 || !strings.Contains(body, "Protected link") {
		t.Fatalf("want PIN page: %d", code)
	}
	if code, _ := get(t, c, base+"/s/"+sh.Token+"/f/0"); code != 401 {
		t.Fatalf("download without PIN: %d", code)
	}
	resp, err := c.PostForm(base+"/s/"+sh.Token+"/pin", url.Values{"pin": {"1234"}})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), "hello.txt") {
		t.Fatal("share page missing file")
	}
	if code, body := get(t, c, base+"/s/"+sh.Token+"/f/0"); code != 200 || body != "hello world" {
		t.Fatalf("%d %q", code, body)
	}
	_, z := get(t, c, base+"/s/"+sh.Token+"/zip")
	zr, err := zip.NewReader(strings.NewReader(z), int64(len(z)))
	if err != nil || len(zr.File) != 2 {
		t.Fatalf("zip: %v", err)
	}
	if got := s.Shares()[0].Downloads; got != 2 {
		t.Fatalf("downloads %d", got)
	}
	s.Remove(sh.Token)
	if code, _ := get(t, c, base+"/s/"+sh.Token); code != http.StatusGone {
		t.Fatal(code)
	}
}

func TestDropUpload(t *testing.T) {
	_, base, dl := start(t)
	resp, err := http.Post(base+"/r/droptok/begin", "application/json", strings.NewReader(`{"files":[{"name":"../a.txt","size":3}]}`))
	if err != nil {
		t.Fatal(err)
	}
	var out struct{ Batch string }
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	out.Batch = strings.Split(strings.Split(string(b), `"batch":"`)[1], `"`)[0]
	resp, err = http.Post(base+"/r/droptok/file?batch="+out.Batch+"&i=0", "application/octet-stream", bytes.NewReader([]byte("abc")))
	if err != nil || resp.StatusCode != 200 {
		t.Fatal(err, resp.Status)
	}
	got, err := os.ReadFile(filepath.Join(dl, "a.txt"))
	if err != nil || string(got) != "abc" {
		t.Fatal(err, string(got))
	}
	if code, _ := get(t, http.DefaultClient, base+"/r/wrong"); code != http.StatusGone {
		t.Fatal(code)
	}
}
