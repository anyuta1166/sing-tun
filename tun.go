package tun

import (
	"io"
	"net"
	"net/netip"
	"runtime"
	"strconv"
	"strings"

	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/ranges"
)

type Handler interface {
	PrepareConnection(network string, source M.Socksaddr, destination M.Socksaddr) error
	N.TCPConnectionHandlerEx
	N.UDPConnectionHandlerEx
}

type Tun interface {
	io.ReadWriter
	N.VectorisedWriter
	Close() error
}

type WinTun interface {
	Tun
	ReadPacket() ([]byte, func(), error)
}

type LinuxTUN interface {
	Tun
	N.FrontHeadroom
	BatchSize() int
	BatchRead(buffers [][]byte, offset int, readN []int) (n int, err error)
	BatchWrite(buffers [][]byte, offset int) error
	TXChecksumOffload() bool
}

const (
	DefaultIPRoute2TableIndex = 2022
	DefaultIPRoute2RuleIndex  = 9000
)

type Options struct {
	Name                     string
	Inet4Address             []netip.Prefix
	Inet6Address             []netip.Prefix
	MTU                      uint32
	GSO                      bool
	AutoRoute                bool
	DNSServers               []netip.Addr
	IPRoute2TableIndex       int
	IPRoute2RuleIndex        int
	AutoRedirectMarkMode     bool
	AutoRedirectInputMark    uint32
	AutoRedirectOutputMark   uint32
	StrictRoute              bool
	Inet4RouteAddress        []netip.Prefix
	Inet6RouteAddress        []netip.Prefix
	Inet4RouteExcludeAddress []netip.Prefix
	Inet6RouteExcludeAddress []netip.Prefix
	RouteMetric              int
	IncludeInterface         []string
	ExcludeInterface         []string
	IncludeUID               []ranges.Range[uint32]
	ExcludeUID               []ranges.Range[uint32]
	IncludeAndroidUser       []int
	IncludePackage           []string
	ExcludePackage           []string
	InterfaceMonitor         DefaultInterfaceMonitor
	FileDescriptor           int
	Logger                   logger.Logger

	// No work for TCP, do not use.
	_TXChecksumOffload bool

	// For library usages.
	EXP_DisableDNSHijack bool
}

func CalculateInterfaceName(name string) (tunName string) {
	if runtime.GOOS == "darwin" {
		tunName = "utun"
	} else if name != "" {
		tunName = name
	} else {
		tunName = "tun"
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}
	var tunIndex int
	for _, netInterface := range interfaces {
		if strings.HasPrefix(netInterface.Name, tunName) {
			index, parseErr := strconv.ParseInt(netInterface.Name[len(tunName):], 10, 16)
			if parseErr == nil && int(index) >= tunIndex {
				tunIndex = int(index) + 1
			}
		}
	}
	tunName = F.ToString(tunName, tunIndex)
	return
}
