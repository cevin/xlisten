package xlisten

import (
	"fmt"
	"github.com/cevin/xlisten/xnet"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const (
	FdNumStart    = 3
	ListenFds     = "LISTEN_FDS"
	ListenFdNames = "LISTEN_FDNAMES"
	ListenPid     = "LISTEN_PID"
)

type Addr struct {
	IP   string
	Port int
}

type TCPListenItem struct {
	Fd         uintptr
	FdName     string
	Network    string
	Addr       net.Addr
	ListenOn   string
	ListenPort int
	Listener   net.Listener
}

type UDPListenItem struct {
	Fd         uintptr
	FdName     string
	Network    string
	Addr       net.Addr
	ListenOn   string
	ListenPort string
	Conn       *net.UDPConn
}

var TCPListens []*TCPListenItem

var UDPListens []*UDPListenItem

func init() {
	Fds, err := strconv.Atoi(os.Getenv(ListenFds))
	if err != nil {
		return
	}

	names := strings.Split(os.Getenv(ListenFdNames), ":")

	for fd := FdNumStart; fd < FdNumStart+Fds; fd++ {
		name := fmt.Sprintf("LISTEN_FD_%d", fd)
		offset := fd - FdNumStart
		if offset < len(names) && len(names[offset]) > 0 {
			name = names[offset]
		}
		syscall.CloseOnExec(fd)
		f := os.NewFile(uintptr(fd), name)
		defer f.Close()

		var host string
		var port string

		if ln, err := net.FileListener(f); err == nil {
			// wrap fd
			_ = xnet.WrapFD(int(f.Fd()))
			switch ln.(type) {
			case *net.UnixListener:
				host = ln.(*net.UnixListener).Addr().(*net.UnixAddr).Name
			default:
				host, port, err = net.SplitHostPort(ln.Addr().String())
				if err != nil {
					host = ""
					port = ""
				}
			}
			listenPort, _ := strconv.Atoi(port)
			TCPListens = append(TCPListens, &TCPListenItem{
				Fd:         uintptr(fd),
				FdName:     name,
				Addr:       ln.Addr(),
				Network:    ln.Addr().Network(),
				ListenOn:   host,
				ListenPort: listenPort,
				Listener:   ln,
			})
		} else if ln, err := net.FilePacketConn(f); err == nil {
			addr, err := ParseAddress("udp", ln.LocalAddr().String())
			if err == nil {
				host = addr.IP
				port = strconv.Itoa(addr.Port)
			}
			UDPListens = append(UDPListens, &UDPListenItem{
				Fd:         uintptr(fd),
				FdName:     name,
				Addr:       ln.LocalAddr(),
				Network:    ln.LocalAddr().Network(),
				ListenOn:   host,
				ListenPort: port,
				Conn:       ln.(*net.UDPConn),
			})
		}

	}
}

// IsCurrentProcessSystemdOwned checks if the current process was started by Systemd
func IsCurrentProcessSystemdOwned() bool {
	envPid, _ := strconv.Atoi(os.Getenv(ListenPid))

	return envPid == os.Getpid()
}

// IsCurrentProcessStartedBySystemd checks if the process is called by Systemd
func IsCurrentProcessStartedBySystemd() bool {
	return len(TCPListens) > 0
}

func RetrieveFirstTCPListener() net.Listener {
	if IsCurrentProcessStartedBySystemd() {
		return TCPListens[0].Listener
	}
	return nil
}

// Listen creates a new listener or returns an existing one based on network and address
func Listen(network, address string) (net.Listener, error) {
	var listenOn string
	var listenPort int

	if strings.HasPrefix(network, "tcp") {
		addr, err := ParseAddress(network, address)
		if err == nil {
			listenOn = addr.IP
			listenPort = addr.Port
		}
	} else {
		listenOn = address
	}

	for _, item := range TCPListens {
		if strings.HasPrefix(item.Network, network) && (listenOn == item.ListenOn && listenPort == item.ListenPort) {
			return item.Listener, nil
		}
	}

	return xnet.Listen(network, address)
}

func ListenUDP(network, address string) (*net.UDPConn, error) {
	addr, err := ParseAddress(network, address)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(network, "udp") && addr != nil {
		for _, item := range UDPListens {
			if item.ListenOn == addr.IP && item.ListenPort == strconv.Itoa(addr.Port) {
				return item.Conn, nil
			}
		}
	}

	return xnet.ListenUDP(network, &net.UDPAddr{
		IP:   net.ParseIP(addr.IP),
		Port: addr.Port,
	})
}

func TCPFilter(f func(item *TCPListenItem) bool) *TCPListenItem {
	for _, item := range TCPListens {
		if f(item) {
			return item
		}
	}
	return nil
}

func UDPFilter(f func(item *UDPListenItem) bool) *UDPListenItem {
	for _, item := range UDPListens {
		if f(item) {
			return item
		}
	}
	return nil
}

// ParseAddress parses an address string and returns an Addr structure
func ParseAddress(network, address string) (*Addr, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if host == "" {
		switch network {
		case "tcp", "tcp4":
			host = "0.0.0.0"
		case "tcp6":
			host = "::1"
		case "udp", "udp4":
			host = "0.0.0.0"
		case "udp6":
			host = "::1"
		default:
			return nil, fmt.Errorf("unsupported network %s", network)
		}
	}

	p, _ := strconv.Atoi(port)

	return &Addr{
		IP:   host,
		Port: p,
	}, nil
}
