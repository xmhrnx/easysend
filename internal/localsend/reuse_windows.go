//go:build windows

package localsend

import "syscall"

// reuse lets several programs (LocalSend and EasySend on one computer)
// listen for the same multicast port.
func reuse(network, address string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		serr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if err != nil {
		return err
	}
	return serr
}
