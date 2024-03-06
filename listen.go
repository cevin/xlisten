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

type ListenItem struct {
	Fd         uintptr
	FdName     string
	Network    string
	Addr       net.Addr
	ListenOn   string
	ListenPort int
	Listener   net.Listener
}

var Listens []*ListenItem

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

		// wrap fd
		_ = xnet.WrapFD(int(f.Fd()))

		if ln, err := net.FileListener(f); err == nil {

			var host string
			var port string

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
			Listens = append(Listens, &ListenItem{
				Fd:         uintptr(fd),
				FdName:     name,
				Addr:       ln.Addr(),
				Network:    ln.Addr().Network(),
				ListenOn:   host,
				ListenPort: listenPort,
				Listener:   ln,
			})
		}
	}
}

// SystemdOwner check current process
func SystemdOwner() bool {
	envPid, _ := strconv.Atoi(os.Getenv(ListenPid))

	return envPid == os.Getpid()
}

func IsCalledBySystemd() bool {
	return len(Listens) > 0
}

func RetrieveFirstListener() net.Listener {
	if IsCalledBySystemd() {
		return Listens[0].Listener
	}
	return nil
}

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

	for _, item := range Listens {
		if strings.HasPrefix(item.Network, network) && (listenOn == item.ListenOn && listenPort == item.ListenPort) {
			return item.Listener, nil
		}
	}

	return xnet.Listen(network, address)
}

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
			host = "[::1]"
		}
	}

	p, _ := strconv.Atoi(port)

	return &Addr{
		IP:   host,
		Port: p,
	}, nil
}
