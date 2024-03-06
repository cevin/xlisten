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
