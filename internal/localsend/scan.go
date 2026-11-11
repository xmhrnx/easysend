package localsend

import (
	"context"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"easysend/internal/diag"
	"easysend/internal/netutil"
)

// Refresh announces again and scans the local subnets for devices that
// don't answer multicast (common on guest Wi-Fi and Android).
func (n *Node) Refresh(ctx context.Context) {
	n.discovery().announce(true)
	n.scan(ctx)
}

func (n *Node) scan(ctx context.Context) {
	sem := make(chan struct{}, 64)
	var wg sync.WaitGroup
	seen := map[string]bool{}
	start := time.Now()
	var open, found atomic.Int32
	defer func() {
		diag.Logf("scan: done in %s: %d open on port %d, %d added", time.Since(start).Round(time.Millisecond), open.Load(), DefaultPort, found.Load())
	}()
	for _, local := range netutil.LocalNets() {
		block := netutil.ScanRange(local)
		if seen[block.String()] {
			continue
		}
		seen[block.String()] = true
		diag.Logf("scan: %s (this device %s)", block.String(), local.IP)
		for _, target := range netutil.Hosts(block) {
			if target.Equal(local.IP) {
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
				open.Add(1)
				c, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()
				if d, err := n.Register(c, addr, false); err != nil {
					diag.Logf("scan: %s is open but didn't identify: %v", addr, err)
				} else {
					found.Add(1)
					diag.Logf("scan: found %q at %s", d.Alias, d.Addr())
				}
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
