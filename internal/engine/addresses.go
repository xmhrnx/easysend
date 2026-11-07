package engine

import (
	"strconv"

	"easysend/internal/netutil"
)

// Self returns this device's name and addresses.
func (e *Engine) Self() Self {
	info := e.node.Self()
	return Self{Name: info.Alias, IP: netutil.PrimaryIPv4(), Model: e.cfg.DeviceModel, Port: e.node.Port(), WebPort: e.web.Port(), Fingerprint: info.Fingerprint}
}

// LANBase is the browser address of this device on the local network.
func (e *Engine) LANBase() string {
	ip := netutil.PrimaryIPv4()
	if ip == "" {
		ip = "localhost"
	}
	return "http://" + ip + ":" + strconv.Itoa(e.web.Port())
}

// OtherLANBases are browser addresses on this device's other local
// addresses (a second adapter), for when the first doesn't open.
func (e *Engine) OtherLANBases() []string {
	ips := netutil.LocalIPv4()
	var out []string
	for i, ip := range ips {
		if i == 0 {
			continue
		}
		out = append(out, "http://"+ip.String()+":"+strconv.Itoa(e.web.Port()))
		if len(out) == 2 {
			break
		}
	}
	return out
}

// PublicBase is the internet address of the browser pages, or "".
func (e *Engine) PublicBase() string {
	if u := e.Settings.Get().PublicURL; u != "" {
		return trimSlash(u)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.public.State == PublicReady {
		return "http://" + e.public.IP + ":" + strconv.Itoa(e.public.Port)
	}
	return ""
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
