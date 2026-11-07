package localsend

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"
)

// Devices returns the known peers, sorted by name.
func (n *Node) Devices() []Device {
	n.mu.Lock()
	out := make([]Device, 0, len(n.devices))
	for _, d := range n.devices {
		out = append(out, *d)
	}
	n.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Alias < out[j].Alias })
	return out
}

// Device returns the peer with a fingerprint.
func (n *Node) Device(fp string) (Device, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	d, ok := n.devices[fp]
	if !ok {
		return Device{}, false
	}
	return *d, true
}

// ForgetDevices clears the list, before a fresh discovery.
func (n *Node) ForgetDevices() {
	n.mu.Lock()
	for fp, d := range n.devices {
		if !d.Manual {
			delete(n.devices, fp)
		}
	}
	n.mu.Unlock()
	n.devicesChanged()
}

func (n *Node) addDevice(d Device) {
	if d.Fingerprint == "" || d.Fingerprint == n.cfg.Fingerprint {
		return
	}
	d.LastSeen = time.Now()
	n.mu.Lock()
	if old, ok := n.devices[d.Fingerprint]; ok && old.Manual {
		d.Manual = true
	}
	n.devices[d.Fingerprint] = &d
	n.mu.Unlock()
	n.devicesChanged()
}

func (n *Node) devicesChanged() {
	if n.cfg.OnDevices != nil {
		n.cfg.OnDevices()
	}
}

// Register asks the device at addr ("ip" or "ip:port") to identify itself
// and adds it, trying HTTPS, then HTTP.
func (n *Node) Register(ctx context.Context, addr string, manual bool) (Device, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		host, portStr = addr, strconv.Itoa(DefaultPort)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Device{}, fmt.Errorf("bad port %q", portStr)
	}
	var lastErr error
	for _, proto := range []string{"https", "http"} {
		d := Device{IP: host, Info: Info{Port: port, Protocol: proto}}
		info, err := n.register(ctx, d)
		if err != nil {
			lastErr = err
			continue
		}
		d.Info = info
		d.Port, d.Protocol, d.Manual = port, proto, manual
		n.addDevice(d)
		return d, nil
	}
	return Device{}, lastErr
}

func (n *Node) register(ctx context.Context, d Device) (Info, error) {
	var info Info
	err := n.postJSON(ctx, d.url("/register"), n.Self(), &info)
	if err != nil {
		return info, err
	}
	if info.Fingerprint == "" {
		return info, errors.New("not a LocalSend device")
	}
	return info, nil
}
