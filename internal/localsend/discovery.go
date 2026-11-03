package localsend

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/ipv4"

	"easysend/internal/netutil"
)

type discovery struct {
	n      *Node
	pc     *ipv4.PacketConn
	conn   net.PacketConn
	group  *net.UDPAddr
	ifaces []net.Interface
	stop   chan struct{}
	once   sync.Once
}

func startDiscovery(n *Node) *discovery {
	d := &discovery{n: n, stop: make(chan struct{}), group: &net.UDPAddr{IP: net.ParseIP(MulticastGroup), Port: DefaultPort}}
	lc := net.ListenConfig{Control: reuse}
	conn, err := lc.ListenPacket(context.Background(), "udp4", ":"+strconv.Itoa(DefaultPort))
	if err != nil {
		// Without the multicast port we can still announce and scan.
		conn, err = net.ListenPacket("udp4", ":0")
		if err != nil {
			return d
		}
	}
	d.conn = conn
	d.pc = ipv4.NewPacketConn(conn)
	_ = d.pc.SetMulticastLoopback(true)
	_ = d.pc.SetMulticastTTL(4)
	if ifs, err := net.Interfaces(); err == nil {
		for _, ifc := range ifs {
			if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagMulticast == 0 {
				continue
			}
			if err := d.pc.JoinGroup(&ifc, d.group); err == nil {
				d.ifaces = append(d.ifaces, ifc)
			}
		}
	}
	go d.listen()
	go d.loop()
	return d
}

func (d *discovery) close() {
	d.once.Do(func() {
		close(d.stop)
		if d.conn != nil {
			_ = d.conn.Close()
		}
	})
}

func (d *discovery) loop() {
	d.announce(true)
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for i := 0; ; i++ {
		select {
		case <-d.stop:
			return
		case <-t.C:
			d.announce(true)
		case <-time.After(2 * time.Second):
			if i < 2 {
				d.announce(true)
			}
		}
	}
}

// announce multicasts this device's info on every interface.
func (d *discovery) announce(isAnnounce bool) {
	if d.pc == nil || !d.n.discoverable() {
		return
	}
	info := d.n.Self()
	info.Announce = &isAnnounce
	info.Announcement = &isAnnounce
	data, err := json.Marshal(info)
	if err != nil {
		return
	}
	if len(d.ifaces) == 0 {
		_, _ = d.pc.WriteTo(data, nil, d.group)
		return
	}
	for i := range d.ifaces {
		if err := d.pc.SetMulticastInterface(&d.ifaces[i]); err != nil {
			continue
		}
		_, _ = d.pc.WriteTo(data, nil, d.group)
	}
}

func (d *discovery) listen() {
	buf := make([]byte, 64<<10)
	for {
		n, _, src, err := d.pc.ReadFrom(buf)
		if err != nil {
			select {
			case <-d.stop:
				return
			default:
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		var info Info
		if json.Unmarshal(buf[:n], &info) != nil || info.Fingerprint == "" || info.Fingerprint == d.n.cfg.Fingerprint {
			continue
		}
		ua, ok := src.(*net.UDPAddr)
		if !ok {
			continue
		}
		dev := Device{Info: info, IP: ua.IP.String()}
		d.n.addDevice(dev)
		if info.isAnnounce() && d.n.discoverable() {
			go d.answer(dev)
		}
	}
}

// answer tells an announcing device about us: over HTTP as LocalSend
// prefers, falling back to a multicast reply.
func (d *discovery) answer(dev Device) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	if _, err := d.n.register(ctx, dev); err != nil {
		d.announce(false)
	}
}

// Refresh announces again and scans the local subnets for devices that
// don't answer multicast (common on guest Wi-Fi and Android).
func (n *Node) Refresh(ctx context.Context) {
	n.disc.announce(true)
	n.scan(ctx)
}

func (n *Node) scan(ctx context.Context) {
	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup
	seen := map[string]bool{}
	for _, ip := range netutil.LocalIPv4() {
		base := ip.Mask(net.CIDRMask(24, 32))
		if seen[base.String()] || !ip.IsPrivate() {
			continue
		}
		seen[base.String()] = true
		for h := 1; h < 255; h++ {
			target := net.IPv4(base[0], base[1], base[2], byte(h))
			if target.Equal(ip) {
				continue
			}
			select {
			case <-ctx.Done():
				wg.Wait()
				return
			case sem <- struct{}{}:
			}
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				defer func() { <-sem }()
				if !portOpen(ctx, addr) {
					return
				}
				c, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()
				_, _ = n.Register(c, addr, false)
			}(net.JoinHostPort(target.String(), strconv.Itoa(DefaultPort)))
		}
	}
	wg.Wait()
}

func portOpen(ctx context.Context, addr string) bool {
	dl := net.Dialer{Timeout: 400 * time.Millisecond}
	c, err := dl.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}
