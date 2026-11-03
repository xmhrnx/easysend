// Package netutil finds this device's local network addresses.
package netutil

import (
	"net"
	"sort"
	"strings"
)

// virtualNames are parts of the names of adapters other devices on the
// network can't reach: VMs, containers, VPNs and tunnels.
var virtualNames = []string{
	"docker", "veth", "br-", "virbr", "vbox", "virtualbox", "vmnet", "vmware", "vethernet", "hyper-v", "wsl",
	"default switch", "tun", "tap", "utun", "zt", "zerotier", "tailscale", "wg", "wireguard", "hamachi",
	"radmin", "npcap", "loopback", "bluetooth", "vpn", "ppp", "ipsec", "llw", "awdl", "anpi", "bridge",
}

// LocalIPv4 returns the device's private IPv4 addresses, the one other
// devices on the local network most likely reach first: the address of
// the adapter that carries the default route, unless that's a VPN or
// virtual adapter, then Wi-Fi and Ethernet before anything virtual.
func LocalIPv4() []net.IP {
	route := routeIPv4()
	ifaces, err := net.Interfaces()
	if err != nil {
		// Android 11+ denies apps the interface list.
		return route
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
			if ip == nil || ip.IsLinkLocalUnicast() || !ip.IsPrivate() {
				continue
			}
			cands = append(cands, cand{ip, score(ifc.Name, ip, route)})
		}
	}
	if len(cands) == 0 {
		return route
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	out := make([]net.IP, len(cands))
	for i, c := range cands {
		out[i] = c.ip
	}
	return out
}

// score ranks an address: the routed adapter first, physical adapters
// next, virtual ones last.
func score(name string, ip net.IP, route []net.IP) int {
	n := strings.ToLower(name)
	if isVirtual(n) {
		return -10
	}
	s := 0
	if len(route) > 0 && route[0].Equal(ip) {
		s += 20
	}
	for _, v := range []string{"wlan", "wi-fi", "wifi", "wireless", "ethernet", "en0", "en1", "eth", "wl", "en"} {
		if strings.HasPrefix(n, v) {
			s += 5
			break
		}
	}
	return s
}

func isVirtual(name string) bool {
	for _, v := range virtualNames {
		if strings.Contains(name, v) {
			return true
		}
	}
	return false
}

// routeIPv4 returns the address the system uses to reach other networks:
// connecting a UDP socket picks the route without sending anything. It
// also works where the interface list is unavailable.
func routeIPv4() []net.IP {
	c, err := net.Dial("udp4", "192.0.2.1:9")
	if err != nil {
		return nil
	}
	defer c.Close()
	ip := c.LocalAddr().(*net.UDPAddr).IP.To4()
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return nil
	}
	return []net.IP{ip}
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
