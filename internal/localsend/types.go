// Package localsend implements the LocalSend v2 protocol, so EasySend
// discovers and exchanges files with other EasySend devices and with the
// LocalSend apps on the local network.
//
// Protocol: https://github.com/localsend/protocol
package localsend

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Protocol constants.
const (
	MulticastGroup = "224.0.0.167"
	DefaultPort    = 53317
	ProtocolVer    = "2.1"
	apiBase        = "/api/localsend/v2"
)

// Device types LocalSend knows.
const (
	TypeMobile   = "mobile"
	TypeDesktop  = "desktop"
	TypeWeb      = "web"
	TypeHeadless = "headless"
	TypeServer   = "server"
)

// Errors returned when sending.
var (
	ErrPINRequired = errors.New("the receiver requires a PIN")
	ErrDeclined    = errors.New("the receiver declined")
	ErrBusy        = errors.New("the receiver is busy with another transfer")
	ErrCanceled    = errors.New("canceled")
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

// Device is a peer found on the network.
type Device struct {
	Info
	IP       string
	LastSeen time.Time
	// Manual is set for devices added by address.
	Manual bool
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
