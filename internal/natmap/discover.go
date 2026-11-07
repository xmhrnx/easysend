package natmap

import (
	"context"
	"net"
	"strings"
	"time"
)

var searchTargets = []string{
	"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
	"urn:schemas-upnp-org:device:InternetGatewayDevice:2",
	"urn:schemas-upnp-org:service:WANIPConnection:1",
}

func discover(ctx context.Context) (*gateway, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	dst := &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900}
	for _, st := range searchTargets {
		msg := "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nST: " + st + "\r\nMAN: \"ssdp:discover\"\r\nMX: 2\r\n\r\n"
		_, _ = conn.WriteTo([]byte(msg), dst)
	}
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetReadDeadline(deadline)
	buf := make([]byte, 4096)
	tried := map[string]bool{}
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			return nil, ErrNoGateway
		}
		loc := header(string(buf[:n]), "location")
		if loc == "" || tried[loc] {
			continue
		}
		tried[loc] = true
		if gw, err := describe(ctx, loc); err == nil {
			return gw, nil
		}
	}
}

func header(resp, name string) string {
	for _, line := range strings.Split(resp, "\r\n") {
		k, v, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(k), name) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
