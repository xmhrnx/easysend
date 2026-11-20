package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"easysend/internal/foldersync"
	"easysend/internal/item"
	"easysend/internal/localsend"
)

// SyncedFolder is a synced folder as the UI shows it.
type SyncedFolder = foldersync.View

// Errors from InviteToSync the UI explains.
var (
	ErrSyncNotSupported = foldersync.ErrNotEasySend
	ErrSyncDeclined     = foldersync.ErrDeclined
)

// SyncFolders lists the synced folders, by name.
func (e *Engine) SyncFolders() []SyncedFolder { return e.sync.Folders() }

// AddSyncFolder starts syncing the folder at root; InviteToSync shares it
// with devices.
func (e *Engine) AddSyncFolder(name, root string) (SyncedFolder, error) {
	f, err := e.sync.Add(name, root)
	if err != nil {
		return SyncedFolder{}, err
	}
	e.changed()
	return SyncedFolder{Folder: f}, nil
}

// SyncFolderDir is where a folder made on this device for syncing goes
// when the user can't pick one (phones): a new folder in the received
// files' folder.
func (e *Engine) SyncFolderDir(name string) string {
	return freeDir(filepath.Join(e.Settings.Get().DownloadDir, "Synced"), name)
}

// InviteToSync offers a folder to a nearby device and waits for its
// user's answer: ErrSyncDeclined, or ErrSyncNotSupported for a device
// without folder sync.
func (e *Engine) InviteToSync(ctx context.Context, id string, p Peer) error {
	if p.Protocol == "http" {
		return ErrSyncNotSupported // folder sync needs device certificates
	}
	err := e.sync.Invite(ctx, id, p.Fingerprint, p.Alias, p.BaseURL())
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return errors.New("no answer from " + p.Alias)
	}
	return err
}

// RemoveSyncFolder stops syncing a folder; its files stay.
func (e *Engine) RemoveSyncFolder(id string) { e.sync.Remove(id) }

// StopSyncWith stops syncing a folder with one device.
func (e *Engine) StopSyncWith(id, fp string) { e.sync.RemovePeer(id, fp) }

// SyncNow syncs every folder now.
func (e *Engine) SyncNow() { e.sync.SyncNow() }

func (e *Engine) resolveDevice(fp string) (string, bool) {
	d, ok := e.node.Device(fp)
	if !ok || d.Protocol == "http" {
		return "", false
	}
	return d.BaseURL(), true
}

// onSyncInvite asks the user about a folder a device offers to sync,
// through the same queue as incoming files - but never accepted
// automatically, since it writes into a folder for good.
func (e *Engine) onSyncInvite(ctx context.Context, inv foldersync.Invite) (string, bool) {
	sender, ok := e.node.Device(inv.FP)
	if !ok {
		sender = localsend.Device{Info: localsend.Info{Alias: inv.From, Fingerprint: inv.FP}}
	}
	if sender.Alias == "" {
		sender.Alias = "A device"
	}
	req := &Request{
		ID: item.NewID(), Sender: sender, Received: time.Now(), answer: make(chan bool, 1),
		SyncFolder: inv.Name, SyncRoot: e.SyncFolderDir(inv.Name),
	}
	for _, f := range e.sync.Folders() {
		if f.ID == inv.ID {
			req.SyncRoot = f.Root // already synced here, with other devices
		}
	}
	if !e.ask(ctx, req, sender.Alias+" wants to sync a folder", fmt.Sprintf("%q, kept in sync on both devices", inv.Name)) {
		return "", false
	}
	e.mu.Lock()
	root := req.SyncRoot
	e.mu.Unlock()
	return root, true
}

// freeDir is dir/name, or dir/"name 2", ... when that's taken by
// something other than an empty folder.
func freeDir(dir, name string) string {
	name = filepath.Base(item.SafeJoin("", name))
	for i := 1; ; i++ {
		n := name
		if i > 1 {
			n = fmt.Sprintf("%s %d", name, i)
		}
		p := filepath.Join(dir, n)
		entries, err := os.ReadDir(p)
		if os.IsNotExist(err) || (err == nil && len(entries) == 0) {
			return p
		}
	}
}

// SyncPeer is a device a folder is synced with.
type SyncPeer = foldersync.Peer

// SyncFolderInfo is a synced folder without its status.
type SyncFolderInfo = foldersync.Folder
