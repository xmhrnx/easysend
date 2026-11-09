// Package diag keeps a short log of what the networking did - ports bound,
// adapters, multicast, scans, link requests - for problem reports. Errors
// that are handled by falling back (a taken port, an adapter that can't
// join multicast) are recorded here so a report shows them.
package diag

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const keep = 400

var (
	mu    sync.Mutex
	lines []string
)

// Logf records a line.
func Logf(format string, args ...any) {
	line := time.Now().Format("15:04:05.000 ") + fmt.Sprintf(format, args...)
	mu.Lock()
	lines = append(lines, line)
	if len(lines) > keep {
		lines = append([]string(nil), lines[len(lines)-keep:]...)
	}
	mu.Unlock()
}

// Lines returns the recorded lines, oldest first.
func Lines() string {
	mu.Lock()
	defer mu.Unlock()
	return strings.Join(lines, "\n")
}
