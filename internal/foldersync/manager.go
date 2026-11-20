package foldersync

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

// Config sets up a Manager.
type Config struct {
	// Dir keeps the folder list and each folder's index.
	Dir string
	// Self is this device's certificate fingerprint.
	Self string
	// Name is this device's name, shown to peers and in conflict copies.
	Name func() string
	// Client returns an HTTP client that only talks to the device with
	// fingerprint fp and presents this device's certificate.
	Client func(fp string) *http.Client
	// Resolve returns a device's current address ("https://ip:port"),
	// from discovery, or false when it isn't around.
	Resolve func(fp string) (string, bool)
	// PeerFP returns the fingerprint of the device behind a request, ""
	// when it presented no certificate.
	PeerFP func(r *http.Request) string
	// OnInvite asks the user whether to sync a folder a device offers,
	// blocking until they answer, and returns where to keep it.
	OnInvite func(ctx context.Context, inv Invite) (root string, ok bool)
	// OnChange runs after anything the UI shows changed.
	OnChange func()
	// Interval is how often folders are scanned and peers asked for
	// changes; 10 seconds by default. Peers also tell each other about
	// changes at once.
	Interval time.Duration
}

// Invite is a device offering to sync a folder.
type Invite struct {
	FP, From string // the device: fingerprint and name
	ID, Name string // the folder
}

// Status is how a folder's syncing is going.
type Status struct {
	Syncing bool
	// Done and Total count the files being fetched or deleted in the
	// current round.
	Done, Total int
	LastSync    time.Time
	Err         string
	// Files and Bytes are what the folder holds on this device.
	Files int
	Bytes int64
	Peers map[string]PeerStatus
}

// PeerStatus is how syncing with one device is going.
type PeerStatus struct {
	Online   bool
	LastSync time.Time
	Err      string
}

// View is a folder with its status, for the UI.
type View struct {
	Folder
	Status Status
}

type folder struct {
	Folder
	st     state
	status Status
	run    sync.Mutex // one sync round at a time
	// asked holds peers that announced changes since the last round.
	asked map[string]bool
}

// Manager syncs the folders of one device.
type Manager struct {
	cfg     Config
	mu      sync.Mutex
	folders map[string]*folder
	wake    chan struct{}
	stop    chan struct{}
	done    chan struct{}
}

// New loads the folders kept in cfg.Dir. Start begins syncing them.
func New(cfg Config) (*Manager, error) {
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}
	m := &Manager{cfg: cfg, folders: map[string]*folder{}, wake: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{})}
	if err := m.load(); err != nil {
		return nil, fmt.Errorf("synced folders: %w", err)
	}
	return m, nil
}

// Start syncs the folders in the background until Close.
func (m *Manager) Start() { go m.loop() }

// Close stops syncing.
func (m *Manager) Close() {
	select {
	case <-m.stop:
		return
	default:
	}
	close(m.stop)
	<-m.done
}

// SyncNow asks for a round now, for every folder.
func (m *Manager) SyncNow() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *Manager) loop() {
	defer close(m.done)
	t := time.NewTicker(m.cfg.Interval)
	defer t.Stop()
	m.round()
	for {
		select {
		case <-m.stop:
			return
		case <-t.C:
		case <-m.wake:
		}
		m.round()
	}
}

// round syncs every folder once.
func (m *Manager) round() {
	m.mu.Lock()
	list := make([]*folder, 0, len(m.folders))
	for _, f := range m.folders {
		list = append(list, f)
	}
	m.mu.Unlock()
	for _, f := range list {
		select {
		case <-m.stop:
			return
		default:
		}
		m.syncFolder(f)
	}
}

func (m *Manager) changed() {
	if m.cfg.OnChange != nil {
		m.cfg.OnChange()
	}
}

// Folders lists the synced folders, by name.
func (m *Manager) Folders() []View {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]View, 0, len(m.folders))
	for _, f := range m.folders {
		v := View{Folder: f.Folder, Status: f.status}
		v.Peers = append([]Peer(nil), f.Peers...)
		v.Status.Peers = make(map[string]PeerStatus, len(f.status.Peers))
		for k, s := range f.status.Peers {
			v.Status.Peers[k] = s
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// Add starts syncing the folder at root, shared with no device yet:
// Invite adds them.
func (m *Manager) Add(name, root string) (Folder, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return Folder{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.folders {
		if f.Root == root {
			return Folder{}, fmt.Errorf("%s is already synced as %q", root, f.Name)
		}
	}
	f := &folder{
		Folder: Folder{ID: newID(), Name: name, Root: root, Created: time.Now()},
		st:     state{Local: Index{}, Base: map[string]map[string]string{}},
		status: Status{Peers: map[string]PeerStatus{}},
	}
	m.folders[f.ID] = f
	if err := m.saveFolders(); err != nil {
		delete(m.folders, f.ID)
		return Folder{}, err
	}
	m.SyncNow()
	return f.Folder, nil
}

// ErrNotEasySend is a device that doesn't sync folders: LocalSend, or an
// older EasySend.
var ErrNotEasySend = errors.New("this device doesn't sync folders (it needs EasySend)")

// Invite offers a folder to the device with fingerprint fp at address
// base, waiting until its user answers, and syncs with it from then on.
func (m *Manager) Invite(ctx context.Context, id, fp, name, base string) error {
	m.mu.Lock()
	f, ok := m.folders[id]
	var fname string
	if ok {
		fname = f.Name
	}
	m.mu.Unlock()
	if !ok {
		return errors.New("no such folder")
	}
	if err := m.sendInvite(ctx, fp, base, inviteRequest{ID: id, Name: fname, From: m.name()}); err != nil {
		return err
	}
	m.mu.Lock()
	f.addPeer(Peer{FP: fp, Name: name, Addr: base})
	_ = m.saveFolders()
	m.mu.Unlock()
	m.changed()
	m.SyncNow()
	return nil
}

// Remove stops syncing a folder and tells its devices. The files stay.
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	f, ok := m.folders[id]
	if ok {
		delete(m.folders, id)
		_ = m.saveFolders()
		_ = os.Remove(m.statePath(id))
	}
	m.mu.Unlock()
	if !ok {
		return
	}
	for _, p := range f.Peers {
		go m.sendLeave(f.ID, p)
	}
	m.changed()
}

// RemovePeer stops syncing a folder with one device, and tells it.
func (m *Manager) RemovePeer(id, fp string) {
	m.mu.Lock()
	f, ok := m.folders[id]
	var p Peer
	if ok {
		p, ok = f.removePeer(fp)
		delete(f.st.Base, fp)
		_ = m.saveFolders()
		_ = m.saveState(f)
	}
	m.mu.Unlock()
	if ok {
		go m.sendLeave(id, p)
		m.changed()
	}
}

func (f *folder) addPeer(p Peer) {
	for i, q := range f.Peers {
		if q.FP == p.FP {
			f.Peers[i] = p
			return
		}
	}
	f.Peers = append(f.Peers, p)
}

func (f *folder) removePeer(fp string) (Peer, bool) {
	for i, q := range f.Peers {
		if q.FP == fp {
			f.Peers = append(f.Peers[:i], f.Peers[i+1:]...)
			return q, true
		}
	}
	return Peer{}, false
}

func (f *folder) hasPeer(fp string) bool {
	for _, p := range f.Peers {
		if p.FP == fp {
			return true
		}
	}
	return false
}

func (m *Manager) name() string {
	if m.cfg.Name != nil {
		return m.cfg.Name()
	}
	return "EasySend"
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
