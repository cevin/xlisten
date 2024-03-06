//go:build (darwin || linux || unix) && !plan9

package xnet

import (
	"golang.org/x/sys/unix"
	"syscall"
)

func Control(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		_ = WrapFD(int(fd))
	})
}

func WrapFD(fd int) error {
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return err
	}

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return err
	}

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.TCP_NODELAY, 1); err != nil {
		return err
	}

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_BROADCAST, 1); err != nil {
		return err
	}

	return nil
}
