package netutil

import (
	"net"
	"testing"
)

func TestScore(t *testing.T) {
	route := []net.IP{net.IPv4(192, 168, 1, 20)}
	wifi := score("Wi-Fi", net.IPv4(192, 168, 1, 20), route)
	for _, name := range []string{"VirtualBox Host-Only Network", "vEthernet (WSL)", "VMware Network Adapter VMnet8", "docker0", "tailscale0", "utun3"} {
		if s := score(name, net.IPv4(192, 168, 56, 1), route); s >= wifi {
			t.Errorf("%s scored %d, Wi-Fi %d", name, s, wifi)
		}
	}
	if eth := score("Ethernet", net.IPv4(10, 0, 0, 5), route); eth >= wifi {
		t.Errorf("unrouted Ethernet %d >= routed Wi-Fi %d", eth, wifi)
	}
}

func TestScanRange(t *testing.T) {
	cases := []struct {
		ip     string
		prefix int
		want   string
		hosts  int
	}{
		{"192.168.1.20", 24, "192.168.1.0/24", 254},
		{"192.168.1.20", 23, "192.168.0.0/23", 510}, // phone at 192.168.0.x is now scanned
		{"10.0.5.20", 16, "10.0.4.0/22", 1022},      // large networks: a /22 around this device
		{"172.16.3.9", 26, "172.16.3.0/26", 62},
	}
	for _, c := range cases {
		n := net.IPNet{IP: net.ParseIP(c.ip).To4(), Mask: net.CIDRMask(c.prefix, 32)}
		r := ScanRange(n)
		if r.String() != c.want {
			t.Errorf("%s/%d: range %s, want %s", c.ip, c.prefix, r.String(), c.want)
		}
		if h := Hosts(r); len(h) != c.hosts {
			t.Errorf("%s: %d hosts, want %d", r.String(), len(h), c.hosts)
		}
	}
	if h := Hosts(ScanRange(net.IPNet{IP: net.IPv4(192, 168, 0, 7).To4(), Mask: net.CIDRMask(23, 32)})); !h[0].Equal(net.IPv4(192, 168, 0, 1)) || !h[len(h)-1].Equal(net.IPv4(192, 168, 1, 254)) {
		t.Errorf("bounds %v .. %v", h[0], h[len(h)-1])
	}
}
