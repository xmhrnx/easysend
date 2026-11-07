package main

import (
	"log"
	"sync"

	"github.com/xmhrnx/giode/multicast"
)

// multicastLock lets Android deliver the multicast packets nearby
// discovery listens for, while the app is on screen (it costs battery).
// Elsewhere it does nothing.
type multicastLock struct {
	mu      sync.Mutex
	release func()
}

func (m *multicastLock) start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.release != nil {
		return
	}
	release, err := multicast.Acquire("EasySend discovery")
	if err != nil {
		log.Printf("multicast: %v", err)
		return
	}
	m.release = release
}

func (m *multicastLock) stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.release != nil {
		m.release()
		m.release = nil
	}
}
