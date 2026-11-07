package localsend

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Info is a device's self description, as exchanged on the wire.
type Info struct {
	Alias       string  `json:"alias"`
	Version     string  `json:"version"`
	DeviceModel *string `json:"deviceModel"`
	DeviceType  string  `json:"deviceType"`
	Fingerprint string  `json:"fingerprint"`
	Port        int     `json:"port,omitempty"`
	Protocol    string  `json:"protocol,omitempty"`
	Download    bool    `json:"download"`
	Announce    *bool   `json:"announce,omitempty"`
	// Announcement is the v1 spelling of Announce.
	Announcement *bool `json:"announcement,omitempty"`
}

// Device is a peer found on the network.
type Device struct {
	Info
	IP       string
	LastSeen time.Time
	// Manual is set for devices added by address.
	Manual bool
}

// FileMeta describes one offered file.
type FileMeta struct {
	ID       string    `json:"id"`
	FileName string    `json:"fileName"`
	Size     int64     `json:"size"`
	FileType string    `json:"fileType"`
	SHA256   string    `json:"sha256,omitempty"`
	Preview  string    `json:"preview,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// Metadata carries optional file times.
type Metadata struct {
	Modified string `json:"modified,omitempty"`
	Accessed string `json:"accessed,omitempty"`
}

type prepareRequest struct {
	Info  Info                `json:"info"`
	Files map[string]FileMeta `json:"files"`
}

type prepareResponse struct {
	SessionID string            `json:"sessionId"`
	Files     map[string]string `json:"files"`
}

// Model returns the device model or "".
func (i Info) Model() string {
	if i.DeviceModel == nil {
		return ""
	}
	return *i.DeviceModel
}

func (i Info) isAnnounce() bool {
	return (i.Announce != nil && *i.Announce) || (i.Announcement != nil && *i.Announcement)
}

// Addr is host:port.
func (d Device) Addr() string { return net.JoinHostPort(d.IP, strconv.Itoa(d.port())) }

func (d Device) port() int {
	if d.Port == 0 {
		return DefaultPort
	}
	return d.Port
}

func (d Device) scheme() string {
	if d.Protocol == "http" {
		return "http"
	}
	return "https"
}

func (d Device) url(path string) string {
	return fmt.Sprintf("%s://%s%s%s", d.scheme(), d.Addr(), apiBase, path)
}
