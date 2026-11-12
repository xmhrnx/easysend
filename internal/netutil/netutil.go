// Package netutil finds this device's local network addresses.
package netutil

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
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
	nets := LocalNets()
	out := make([]net.IP, len(nets))
	for i, n := range nets {
		out[i] = n.IP
	}
	return out
}

// netsTTL is how long LocalNets reuses what it found: the UI asks for
// addresses every frame (links, QR codes, this device's card), and
// listing the interfaces is slow where there are many (an iPhone has
// twenty-odd). Network changes still show within it.
const netsTTL = 2 * time.Second

var netsCache struct {
	sync.Mutex
	at   time.Time
	nets []net.IPNet
}

// LocalNets is LocalIPv4 with each address's network (its subnet mask),
// for scanning the network. Where the system doesn't say (Android 11+
// denies apps the interface list) the network is assumed to be a /24.
// Results are reused for netsTTL.
func LocalNets() []net.IPNet {
	netsCache.Lock()
	defer netsCache.Unlock()
	if netsCache.nets != nil && time.Since(netsCache.at) < netsTTL {
		return append([]net.IPNet(nil), netsCache.nets...)
	}
	netsCache.nets, netsCache.at = localNets(), time.Now()
	if netsCache.nets == nil {
		netsCache.nets = []net.IPNet{} // cached too: no network is an answer
	}
	return append([]net.IPNet(nil), netsCache.nets...)
}

func localNets() []net.IPNet {
	route := routeIPv4()
	ifaces, err := net.Interfaces()
	if err != nil {
		return assumed24(route)
	}
	type cand struct {
		net   net.IPNet
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
			mask := ipn.Mask
			if len(mask) == net.IPv6len {
				mask = mask[12:]
			}
			cands = append(cands, cand{net.IPNet{IP: ip, Mask: mask}, score(ifc.Name, ip, route)})
		}
	}
	if len(cands) == 0 {
		return assumed24(route)
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	out := make([]net.IPNet, len(cands))
	for i, c := range cands {
		out[i] = c.net
	}
	return out
}

func assumed24(ips []net.IP) []net.IPNet {
	out := make([]net.IPNet, len(ips))
	for i, ip := range ips {
		out[i] = net.IPNet{IP: ip, Mask: net.CIDRMask(24, 32)}
	}
	return out
}

// ScanRange is the block of addresses to look for devices in around n:
// its whole subnet, but at most a /22 (1022 hosts) around n.IP, so a
// large network still scans in seconds.
func ScanRange(n net.IPNet) net.IPNet {
	ones, bits := n.Mask.Size()
	if bits != 32 || ones < 22 {
		ones = 22
	}
	if ones > 30 {
		ones = 30
	}
	mask := net.CIDRMask(ones, 32)
	return net.IPNet{IP: n.IP.Mask(mask), Mask: mask}
}

// Hosts lists the usable addresses of a block, without its network and
// broadcast addresses.
func Hosts(n net.IPNet) []net.IP {
	base := n.IP.To4()
	ones, _ := n.Mask.Size()
	count := uint32(1) << (32 - ones)
	start := uint32(base[0])<<24 | uint32(base[1])<<16 | uint32(base[2])<<8 | uint32(base[3])
	out := make([]net.IP, 0, count)
	for i := uint32(1); i+1 < count; i++ {
		v := start + i
		out = append(out, net.IPv4(byte(v>>24), byte(v>>16), byte(v>>8), byte(v)).To4())
	}
	return out
}

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

// Describe lists every network adapter with its state and addresses, and
// how EasySend ranks it, for problem reports.
func Describe() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "adapters: " + err.Error() + "; routed address: " + fmt.Sprint(routeIPv4())
	}
	route := routeIPv4()
	var b strings.Builder
	fmt.Fprintf(&b, "routed address: %v\n", route)
	for _, ifc := range ifaces {
		addrs, _ := ifc.Addrs()
		var list []string
		for _, a := range addrs {
			list = append(list, a.String())
		}
		state := "down"
		if ifc.Flags&net.FlagUp != 0 {
			state = "up"
		}
		if ifc.Flags&net.FlagMulticast != 0 {
			state += ",multicast"
		}
		if ifc.Flags&net.FlagLoopback != 0 {
			state += ",loopback"
		}
		rank := ""
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok && ipn.IP.To4() != nil {
				rank = fmt.Sprintf(" score=%d", score(ifc.Name, ipn.IP.To4(), route))
				break
			}
		}
		fmt.Fprintf(&b, "%q [%s]%s %s\n", ifc.Name, state, rank, strings.Join(list, " "))
	}
	return b.String()
}
