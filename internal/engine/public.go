package engine

import (
	"context"
	"time"

	"easysend/internal/natmap"
)

// PublicState is where opening the browser pages to the internet stands.
type PublicState int

const (
	// PublicOff means UPnP is turned off.
	PublicOff PublicState = iota
	// PublicWorking means the router is being asked.
	PublicWorking
	// PublicReady means links work from the internet.
	PublicReady
	// PublicUnavailable means the router didn't open a port.
	PublicUnavailable
)

// PublicStatus reports internet reachability of the browser pages.
type PublicStatus struct {
	State PublicState
	IP    string
	Port  int
	Err   string
}

// Public returns the internet reachability of the browser pages.
func (e *Engine) Public() PublicStatus {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.public
}

const leaseTime = time.Hour

// applyPublic opens or closes the router port for the browser pages.
func (e *Engine) applyPublic() {
	e.mu.Lock()
	if e.stopNAT != nil {
		e.stopNAT()
		e.stopNAT = nil
	}
	old := e.mapping
	e.mapping = nil
	on := e.Settings.Get().UPnP
	if on {
		e.public = PublicStatus{State: PublicWorking}
	} else {
		e.public = PublicStatus{State: PublicOff}
	}
	var ctx context.Context
	if on {
		ctx, e.stopNAT = context.WithCancel(context.Background())
	}
	e.mu.Unlock()
	e.changed()
	if old != nil {
		go func() {
			c, cl := context.WithTimeout(context.Background(), 3*time.Second)
			defer cl()
			_ = old.Remove(c)
		}()
	}
	if !on {
		return
	}
	go e.keepMapped(ctx)
}

func (e *Engine) keepMapped(ctx context.Context) {
	const desc = "EasySend browser sharing"
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	m, err := natmap.Map(c, e.web.Port(), desc, leaseTime)
	cancel()
	e.mu.Lock()
	switch {
	case ctx.Err() != nil:
	case err != nil:
		e.public = PublicStatus{State: PublicUnavailable, Err: "Your router didn't open a port (UPnP is off or unsupported)."}
	case !m.IsPublic():
		e.public = PublicStatus{State: PublicUnavailable, IP: m.ExternalIP, Err: "Your router is itself behind another network (carrier NAT), so it can't be reached from the internet."}
	default:
		e.mapping = m
		e.public = PublicStatus{State: PublicReady, IP: m.ExternalIP, Port: m.ExternalPort}
	}
	e.mu.Unlock()
	e.changed()
	if err != nil || ctx.Err() != nil || !m.IsPublic() {
		if m != nil && !m.IsPublic() {
			c, cl := context.WithTimeout(context.Background(), 3*time.Second)
			_ = m.Remove(c)
			cl()
		}
		return
	}
	t := time.NewTicker(leaseTime / 2)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c, cl := context.WithTimeout(ctx, 10*time.Second)
			_ = m.Renew(c, leaseTime, desc)
			cl()
		}
	}
}
