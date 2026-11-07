// Package natmap opens a port on the router with UPnP IGD, so devices on
// the internet can reach a server on this device without manual port
// forwarding.
package natmap

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"
)

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

// ErrNoGateway means no UPnP router answered.
var ErrNoGateway = errors.New("no UPnP router found")

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
