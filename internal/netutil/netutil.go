// Package netutil finds this device's local network addresses.
package netutil

import (
	"net"
	"sort"
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
