package localsend

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/ipv4"

	"easysend/internal/diag"
)

type discovery struct {
	n      *Node
	pc     *ipv4.PacketConn
	conn   net.PacketConn
	group  *net.UDPAddr
	ifaces []net.Interface
	stop   chan struct{}
	once   sync.Once
	logged sync.Map // keys of lines logged once
}

func startDiscovery(n *Node) *discovery {
	d := &discovery{n: n, stop: make(chan struct{}), group: &net.UDPAddr{IP: net.ParseIP(MulticastGroup), Port: DefaultPort}}
	lc := net.ListenConfig{Control: reuse}
	conn, err := lc.ListenPacket(context.Background(), "udp4", ":"+strconv.Itoa(DefaultPort))
	if err != nil {
		// Without the multicast port we can still announce and scan, but
		// won't hear other devices' announcements.
		diag.Logf("discovery: UDP port %d unavailable (%v): won't hear announcements", DefaultPort, err)
		conn, err = net.ListenPacket("udp4", ":0")
		if err != nil {
			diag.Logf("discovery: no UDP socket: %v", err)
			return d
		}
	} else {
		diag.Logf("discovery: listening on UDP %d", DefaultPort)
	}
	d.conn = conn
	d.pc = ipv4.NewPacketConn(conn)
	_ = d.pc.SetMulticastLoopback(true)
	_ = d.pc.SetMulticastTTL(4)
	ifs, err := net.Interfaces()
	if err != nil {
		diag.Logf("discovery: can't list adapters (%v): announcing on the default one", err)
	}
	for _, ifc := range ifs {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagMulticast == 0 {
			continue
		}
		if err := d.pc.JoinGroup(&ifc, d.group); err != nil {
			diag.Logf("discovery: %q can't join multicast: %v", ifc.Name, err)
			continue
		}
		d.ifaces = append(d.ifaces, ifc)
		diag.Logf("discovery: joined multicast on %q", ifc.Name)
	}
	go d.listen()
	go d.loop()
	return d
}

// logOnce records a line the first time key comes up, so repeating
// events (every announcement) don't fill the log.
func (d *discovery) logOnce(key, format string, args ...any) {
	if _, seen := d.logged.LoadOrStore(key, true); !seen {
		diag.Logf(format, args...)
	}
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
		if _, err := d.pc.WriteTo(data, nil, d.group); err != nil {
			d.logOnce("announce", "discovery: announcing failed: %v", err)
		}
		return
	}
	for i := range d.ifaces {
		if err := d.pc.SetMulticastInterface(&d.ifaces[i]); err != nil {
			d.logOnce("if-"+d.ifaces[i].Name, "discovery: can't announce on %q: %v", d.ifaces[i].Name, err)
			continue
		}
		if _, err := d.pc.WriteTo(data, nil, d.group); err != nil {
			d.logOnce("send-"+d.ifaces[i].Name, "discovery: announcing on %q failed: %v", d.ifaces[i].Name, err)
		}
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
		d.logOnce("heard-"+info.Fingerprint, "discovery: heard %q (%s) at %s:%d over multicast", info.Alias, info.DeviceType, dev.IP, info.Port)
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
		d.logOnce("answer-"+dev.Fingerprint, "discovery: can't reach %q at %s to answer (%v); answering over multicast", dev.Alias, dev.Addr(), err)
		d.announce(false)
	}
}
