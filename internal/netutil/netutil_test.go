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
