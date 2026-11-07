package localsend

import (
	"context"
	"net"
	"strconv"
	"sync"
	"time"

	"easysend/internal/netutil"
)

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
