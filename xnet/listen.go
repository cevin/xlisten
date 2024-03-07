package xnet

import (
	"context"
	"net"
)

func Listen(network, address string) (net.Listener, error) {
	netListenerConfig := &net.ListenConfig{
		Control: Control,
	}

	return netListenerConfig.Listen(context.TODO(), network, address)
}

func ListenUDP(network string, addr *net.UDPAddr) (*net.UDPConn, error) {
	return net.ListenUDP(network, addr)
}
