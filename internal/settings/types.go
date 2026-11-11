package settings

// AutoAccept says whose incoming nearby transfers are saved without asking.
type AutoAccept int

const (
	// AcceptAsk always asks.
	AcceptAsk AutoAccept = iota
	// AcceptFavorites saves transfers from favorite devices without asking.
	AcceptFavorites
	// AcceptEveryone saves every transfer without asking.
	AcceptEveryone
)

// Theme is the color mode preference.
type Theme int

const (
	// ThemeSystem follows the platform.
	ThemeSystem Theme = iota
	// ThemeDark is always dark.
	ThemeDark
	// ThemeLight is always light.
	ThemeLight
)

// Settings are the user's preferences.
type Settings struct {
	DeviceName  string `json:"deviceName"`
	Fingerprint string `json:"fingerprint"`
	Theme       Theme  `json:"theme"`
	Accent      int    `json:"accent"`

	DownloadDir   string     `json:"downloadDir"`
	AutoAccept    AutoAccept `json:"autoAccept"`
	RequirePIN    bool       `json:"requirePin"`
	PIN           string     `json:"pin"`
	Discoverable  bool       `json:"discoverable"`
	Notifications bool       `json:"notifications"`
	// KeepAfterSend keeps the send list once its items are sent; by
	// default they leave it, so the next send starts empty.
	KeepAfterSend bool `json:"keepAfterSend"`
	// NoBackground stops receiving once the app leaves the screen on a
	// phone; by default it goes on (Android: a foreground service; iOS:
	// the short time iOS allows).
	NoBackground bool `json:"noBackground"`

	MinimizeToTray bool `json:"minimizeToTray"`
	CloseToTray    bool `json:"closeToTray"`
	StartHidden    bool `json:"startHidden"`

	Port        int    `json:"port"`
	WebPort     int    `json:"webPort"`
	UPnP        bool   `json:"upnp"`
	PublicURL   string `json:"publicUrl"`
	BrowserDrop bool   `json:"browserDrop"`
	DropPIN     string `json:"dropPin"`
	DropToken   string `json:"dropToken"`

	RendezvousURL string `json:"rendezvousUrl"`
	TransitRelay  string `json:"transitRelay"`

	// Favorites maps device fingerprints to their last known alias.
	Favorites map[string]string `json:"favorites"`
	// Manual lists addresses ("192.168.1.20:53317") added by hand.
	Manual []string `json:"manual"`
}
