// Package netutil finds this device's local network addresses.
package netutil

import (
	"net"
	"sort"
	"strings"
)

// LocalIPv4 returns the device's private IPv4 addresses, most likely LAN
// address first (Wi-Fi/Ethernet over virtual adapters).
func LocalIPv4() []net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	type cand struct {
		ip    net.IP
		score int
	}
	var cands []cand
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipn.IP.To4()
			if ip == nil || ip.IsLinkLocalUnicast() {
				continue
			}
			cands = append(cands, cand{ip, score(ifc.Name, ip)})
		}
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	out := make([]net.IP, len(cands))
	for i, c := range cands {
		out[i] = c.ip
	}
	return out
}

func score(name string, ip net.IP) int {
	s := 0
	if ip.IsPrivate() {
		s += 10
	}
	if ip[0] == 192 && ip[1] == 168 {
		s += 3
	}
	n := strings.ToLower(name)
	for _, v := range []string{"docker", "veth", "br-", "vbox", "vmnet", "virbr", "tun", "tap", "utun", "zt", "tailscale", "wg"} {
		if strings.Contains(n, v) {
			s -= 8
		}
	}
	for _, v := range []string{"wlan", "wi-fi", "wifi", "en0", "eth", "en", "wl"} {
		if strings.HasPrefix(n, v) {
			s += 2
			break
		}
	}
	return s
}

// PrimaryIPv4 returns the most likely LAN address, or "".
func PrimaryIPv4() string {
	ips := LocalIPv4()
	if len(ips) == 0 {
		return ""
	}
	return ips[0].String()
}

// IsLocal reports whether ip is one of this device's addresses.
func IsLocal(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	for _, l := range LocalIPv4() {
		if l.Equal(ip) {
			return true
		}
	}
	return false
}
