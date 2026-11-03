// Package natmap opens a port on the router with UPnP IGD, so devices on
// the internet can reach a server on this device without manual port
// forwarding.
package natmap

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrNoGateway means no UPnP router answered.
var ErrNoGateway = errors.New("no UPnP router found")

// Mapping is an open port on the router.
type Mapping struct {
	ExternalIP   string
	ExternalPort int
	InternalPort int
	gw           *gateway
}

type gateway struct {
	controlURL  string
	serviceType string
	localIP     string
}

var searchTargets = []string{
	"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
	"urn:schemas-upnp-org:device:InternetGatewayDevice:2",
	"urn:schemas-upnp-org:service:WANIPConnection:1",
}

// Map forwards external TCP port port on the router to port on this
// device for lease (0: until removed), returning the public address.
func Map(ctx context.Context, port int, desc string, lease time.Duration) (*Mapping, error) {
	gw, err := discover(ctx)
	if err != nil {
		return nil, err
	}
	m := &Mapping{InternalPort: port, ExternalPort: port, gw: gw}
	if err := m.add(ctx, lease, desc); err != nil {
		// Some routers only accept permanent leases.
		if err2 := m.add(ctx, 0, desc); err2 != nil {
			return nil, err
		}
	}
	ip, err := gw.externalIP(ctx)
	if err != nil {
		_ = m.Remove(ctx)
		return nil, err
	}
	m.ExternalIP = ip
	return m, nil
}

// Renew extends the lease.
func (m *Mapping) Renew(ctx context.Context, lease time.Duration, desc string) error {
	return m.add(ctx, lease, desc)
}

func (m *Mapping) add(ctx context.Context, lease time.Duration, desc string) error {
	args := []kv{
		{"NewRemoteHost", ""},
		{"NewExternalPort", strconv.Itoa(m.ExternalPort)},
		{"NewProtocol", "TCP"},
		{"NewInternalPort", strconv.Itoa(m.InternalPort)},
		{"NewInternalClient", m.gw.localIP},
		{"NewEnabled", "1"},
		{"NewPortMappingDescription", desc},
		{"NewLeaseDuration", strconv.Itoa(int(lease.Seconds()))},
	}
	_, err := m.gw.soap(ctx, "AddPortMapping", args)
	return err
}

// Remove closes the port.
func (m *Mapping) Remove(ctx context.Context) error {
	_, err := m.gw.soap(ctx, "DeletePortMapping", []kv{
		{"NewRemoteHost", ""},
		{"NewExternalPort", strconv.Itoa(m.ExternalPort)},
		{"NewProtocol", "TCP"},
	})
	return err
}

// IsPublic reports whether the router's external address is reachable
// from the internet (not itself behind carrier-grade NAT).
func (m *Mapping) IsPublic() bool {
	ip := net.ParseIP(m.ExternalIP)
	if ip == nil {
		return false
	}
	cgnat := &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}
	return !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsUnspecified() && !cgnat.Contains(ip)
}

func discover(ctx context.Context) (*gateway, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	dst := &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900}
	for _, st := range searchTargets {
		msg := "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nST: " + st + "\r\nMAN: \"ssdp:discover\"\r\nMX: 2\r\n\r\n"
		_, _ = conn.WriteTo([]byte(msg), dst)
	}
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetReadDeadline(deadline)
	buf := make([]byte, 4096)
	tried := map[string]bool{}
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			return nil, ErrNoGateway
		}
		loc := header(string(buf[:n]), "location")
		if loc == "" || tried[loc] {
			continue
		}
		tried[loc] = true
		if gw, err := describe(ctx, loc); err == nil {
			return gw, nil
		}
	}
}

func header(resp, name string) string {
	for _, line := range strings.Split(resp, "\r\n") {
		k, v, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(k), name) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

type xmlService struct {
	ServiceType string `xml:"serviceType"`
	ControlURL  string `xml:"controlURL"`
}

type xmlDevice struct {
	Services []xmlService `xml:"serviceList>service"`
	Devices  []xmlDevice  `xml:"deviceList>device"`
}

type xmlRoot struct {
	URLBase string    `xml:"URLBase"`
	Device  xmlDevice `xml:"device"`
}

func describe(ctx context.Context, loc string) (*gateway, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loc, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 4 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	gw, err := parseDescription(loc, body)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(loc)
	gw.localIP = localIPFor(u.Host)
	return gw, nil
}

func parseDescription(loc string, body []byte) (*gateway, error) {
	var root xmlRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	svc := findService(root.Device)
	if svc == nil {
		return nil, errors.New("router has no WAN connection service")
	}
	base := root.URLBase
	if base == "" {
		base = loc
	}
	bu, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	cu, err := bu.Parse(svc.ControlURL)
	if err != nil {
		return nil, err
	}
	return &gateway{controlURL: cu.String(), serviceType: svc.ServiceType}, nil
}

func findService(d xmlDevice) *xmlService {
	for i, s := range d.Services {
		if strings.Contains(s.ServiceType, "WANIPConnection") || strings.Contains(s.ServiceType, "WANPPPConnection") {
			return &d.Services[i]
		}
	}
	for _, c := range d.Devices {
		if s := findService(c); s != nil {
			return s
		}
	}
	return nil
}

func localIPFor(host string) string {
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	c, err := net.Dial("udp4", host)
	if err != nil {
		return ""
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}

type kv struct{ k, v string }

func (g *gateway) soap(ctx context.Context, action string, args []kv) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>`)
	fmt.Fprintf(&b, `<u:%s xmlns:u="%s">`, action, g.serviceType)
	for _, a := range args {
		b.WriteString("<" + a.k + ">")
		_ = xml.EscapeText(&b, []byte(a.v))
		b.WriteString("</" + a.k + ">")
	}
	fmt.Fprintf(&b, `</u:%s></s:Body></s:Envelope>`, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.controlURL, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"`+g.serviceType+"#"+action+`"`)
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("router refused %s (%s)", action, resp.Status)
	}
	return body, nil
}

func (g *gateway) externalIP(ctx context.Context) (string, error) {
	body, err := g.soap(ctx, "GetExternalIPAddress", nil)
	if err != nil {
		return "", err
	}
	var env struct {
		IP string `xml:"Body>GetExternalIPAddressResponse>NewExternalIPAddress"`
	}
	if err := xml.Unmarshal(body, &env); err != nil {
		return "", err
	}
	if env.IP == "" {
		return "", errors.New("router didn't report its address")
	}
	return env.IP, nil
}
