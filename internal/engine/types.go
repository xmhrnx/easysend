package engine

import (
	"context"
	"sync"

	"easysend/internal/localsend"
	"easysend/internal/natmap"
	"easysend/internal/settings"
	"easysend/internal/transfer"
	"easysend/internal/webshare"
)

// Config sets up the engine.
type Config struct {
	// DataDir holds settings, history and the TLS certificate.
	DataDir string
	// DownloadDir is the default folder for received files.
	DownloadDir string
	// DeviceModel ("MacBook Pro", "Pixel 8") and DeviceType
	// (localsend.TypeDesktop or localsend.TypeMobile) describe this device.
	DeviceModel string
	DeviceType  string
	// OnChange runs on any goroutine whenever state the UI shows changed.
	OnChange func()
	// Notify shows a system notification.
	Notify func(title, body string)
	// Publish, when set, moves a received file from DownloadDir to where
	// the user finds it (Android's Downloads) and returns its new
	// location. The original is removed once it succeeds.
	Publish func(ctx context.Context, path string) (string, error)
}

// Engine is the running app core.
type Engine struct {
	cfg       Config
	Settings  *settings.Store
	Transfers *transfer.Registry

	node *localsend.Node
	web  *webshare.Server

	mu       sync.Mutex
	requests []*Request
	public   PublicStatus
	mapping  *natmap.Mapping
	scanning bool
	stopNAT  context.CancelFunc
}

// Self describes this device for the UI.
type Self struct {
	Name, IP, Model string
	Port, WebPort   int
	Fingerprint     string
}
