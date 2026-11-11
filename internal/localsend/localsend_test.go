package localsend

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"sync"
	"testing"
	"time"

	"easysend/internal/item"
)

type memSink struct {
	mu    sync.Mutex
	files map[string][]byte
	done  chan error
}

func (m *memSink) Write(ctx context.Context, id string, r io.Reader) error {
	b, err := io.ReadAll(r)
	m.mu.Lock()
	m.files[id] = b
	m.mu.Unlock()
	return err
}
func (m *memSink) Close(err error) { m.done <- err }

func node(t *testing.T, name string, cfg Config) *Node {
	t.Helper()
	dir := t.TempDir()
	cert, fp, err := LoadOrCreateCert(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Self = func() Info { return Info{Alias: name, DeviceType: TypeDesktop} }
	cfg.Cert = &cert
	cfg.Fingerprint = fp
	cfg.Port = -1
	n, err := Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { n.Close() })
	return n
}

func TestSendFilesAndText(t *testing.T) {
	sink := &memSink{files: map[string][]byte{}, done: make(chan error, 1)}
	var gotText string
	recv := node(t, "Receiver", Config{
		PIN:    func() string { return "4242" },
		Accept: func(ctx context.Context, in Incoming) Sink { return sink },
		OnText: func(from Device, text string) { gotText = text + "|" + from.Alias },
	})
	send := node(t, "Sender", Config{})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dev, err := send.Register(ctx, "127.0.0.1:"+strconv.Itoa(recv.Port()), true)
	if err != nil || dev.Alias != "Receiver" || dev.Protocol != "https" {
		t.Fatalf("register: %v %+v", err, dev)
	}
	if _, ok := recv.Device(send.Self().Fingerprint); !ok {
		t.Fatal("receiver didn't learn about sender")
	}

	data := bytes.Repeat([]byte("easysend"), 50000)
	path := t.TempDir() + "/f.bin"
	if err := writeFile(path, data); err != nil {
		t.Fatal(err)
	}
	it, _ := item.FromFile(path)
	items := []item.Item{it}

	if err := send.Send(ctx, dev, items, "", SendEvents{}); !errors.Is(err, ErrPINRequired) {
		t.Fatalf("want PIN error, got %v", err)
	}
	var sent int64
	if err := send.Send(ctx, dev, items, "4242", SendEvents{Progress: func(i int, n int64) { sent += n }}); err != nil {
		t.Fatal(err)
	}
	if err := <-sink.done; err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sink.files[it.ID], data) || sent != int64(len(data)) {
		t.Fatalf("content mismatch %d/%d", len(sink.files[it.ID]), sent)
	}

	if err := send.Send(ctx, dev, []item.Item{item.FromText("hello")}, "4242", SendEvents{}); err != nil {
		t.Fatal(err)
	}
	if gotText != "hello|Sender" {
		t.Fatalf("text %q", gotText)
	}
}

func TestDecline(t *testing.T) {
	recv := node(t, "R", Config{Accept: func(ctx context.Context, in Incoming) Sink { return nil }})
	send := node(t, "S", Config{})
	ctx := context.Background()
	dev, err := send.Register(ctx, "127.0.0.1:"+strconv.Itoa(recv.Port()), true)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/x"
	_ = writeFile(path, []byte("x"))
	it, _ := item.FromFile(path)
	if err := send.Send(ctx, dev, []item.Item{it}, "", SendEvents{}); !errors.Is(err, ErrDeclined) {
		t.Fatalf("got %v", err)
	}
}

// TestResume checks a node answers on the same port after its sockets
// are replaced, as after iOS resumes the app.
func TestResume(t *testing.T) {
	recv := node(t, "Receiver", Config{})
	send := node(t, "Sender", Config{})
	port := recv.Port()
	recv.Resume()
	if recv.Port() != port {
		t.Fatalf("port changed from %d to %d", port, recv.Port())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dev, err := send.Register(ctx, "127.0.0.1:"+strconv.Itoa(recv.Port()), true)
	if err != nil || dev.Alias != "Receiver" {
		t.Fatalf("register after resume: %v %+v", err, dev)
	}
}
