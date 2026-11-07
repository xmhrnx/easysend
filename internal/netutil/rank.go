package netutil

import (
	"net"
	"strings"
)

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
