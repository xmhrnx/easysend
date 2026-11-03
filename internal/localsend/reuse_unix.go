//go:build !windows

package localsend

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// reuse lets several programs (LocalSend and EasySend on one computer)
// listen for the same multicast port.
func reuse(network, address string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		serr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if serr == nil {
			_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		}
	})
	if err != nil {
		return err
	}
	return serr
}
