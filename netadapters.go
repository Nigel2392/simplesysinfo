package simplesysinfo

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

type NetAdapters []*NetAdapterInfo

func (n NetAdapters) Len() int {
	return len(n)
}

func (n NetAdapters) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("NetAdapters: %d\n", len(n)))
	for _, netAdapter := range n {
		writeToBuf(&b, "Name", netAdapter.Name)
		writeToBuf(&b, "IsUp", netAdapter.IsUp)
		writeToBuf(&b, "IsIpv4", netAdapter.IsIpv4)
		writeToBuf(&b, "IP", netAdapter.IP)
		writeToBuf(&b, "MacAddr", netAdapter.MacAddr)
		b.WriteString("\n")
	}
	return b.String()
}

type NetAdapterInfo struct {
	IsUp    bool   `json:"isup"`
	IsIpv4  bool   `json:"isipv4"`
	Name    string `json:"name"`
	IP      string `json:"ip"`
	MacAddr string `json:"macaddr"`
}

func GetNetAdapters() []*NetAdapterInfo {
	interfaces, _ := net.Interfaces()
	netAdapters := make([]*NetAdapterInfo, 0)
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				// Do not include IPV6 addresses
				var ip, _, err = net.ParseCIDR(addr.String())
				if err != nil {
					continue
				}
				var isIpv4 bool = true
				if ip.To4() == nil {
					isIpv4 = false
				}
				var up bool = true
				if interf.Flags&net.FlagUp == 0 {
					up = false
				}
				netAdapter := &NetAdapterInfo{}
				netAdapter.Name = interf.Name
				netAdapter.MacAddr = interf.HardwareAddr.String()
				netAdapter.IP = addr.String()
				netAdapter.IsUp = up
				netAdapter.IsIpv4 = isIpv4
				netAdapters = append(netAdapters, netAdapter)
				// }
			}
		}
	}
	sort.Slice(netAdapters, func(i, j int) bool {
		return netAdapters[i].Name > netAdapters[j].Name && netAdapters[i].IsIpv4 && !netAdapters[j].IsIpv4
	})
	return netAdapters
}
