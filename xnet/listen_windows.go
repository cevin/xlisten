//go:build windows

package xnet

import "syscall"
import "golang.org/x/sys/windows"

func Control(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		_ = WrapFD(int(fd))
	})
}

func WrapFD(fd int) error {
	if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1); err != nil {
		return err
	}
	if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.TCP_NODELAY, 1); err != nil {
		return err
	}
	if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_BROADCAST, 1); err != nil {
		return err
	}

	return nil
}
