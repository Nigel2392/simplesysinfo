package simplesysinfo

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"sort"
	"strconv"
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
		writeToBuf(&b, 1, "Name", netAdapter.Name)
		writeToBuf(&b, 1, "IsUp", netAdapter.IsUp)
		writeToBuf(&b, 1, "IsIpv4", netAdapter.IsIpv4)
		writeToBuf(&b, 1, "IP", netAdapter.IP)
		writeToBuf(&b, 1, "MacAddr", netAdapter.MacAddr)
		b.WriteString("\n")
	}
	return b.String()
}

type NetAdapterInfo struct {
	IsUp    bool     `json:"isup"`
	IsIpv4  bool     `json:"isipv4"`
	Name    string   `json:"name"`
	IP      string   `json:"ip"`
	MacAddr string   `json:"macaddr"`
	Ports   []*Ports `json:"ports"`
}

type Ports struct {
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	ExternalIP string `json:"external_ip"`
	State      string `json:"state"` // LISTENING, ESTABLISHED, TIME_WAIT, CLOSE_WAIT, etc
	PID        int    `json:"pid"`
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
				var ports []*Ports
				if up && isIpv4 {
					ports = make([]*Ports, 0)
					// netstat -ano
					cmd := exec.Command("netstat", "-ano")
					var out bytes.Buffer
					cmd.Stdout = &out
					err := cmd.Run()
					if err != nil {
						log.Fatal(err)
						goto initAdapter
					}
					lines := strings.Split(out.String(), "\r\n")
					for _, line := range lines {
						if strings.Contains(line, ip.String()) {
							var items = strings.Split(line, " ")
							var actualItems []string
							for _, item := range items {
								if item == "" {
									continue
								}
								actualItems = append(actualItems, item)
							}
							if len(actualItems) < 4 {
								continue
							}
							var port = &Ports{}
							var ipPort = strings.Split(actualItems[1], ":")
							if len(ipPort) < 2 {
								continue
							}
							port.Protocol = actualItems[0]
							port.Port, _ = strconv.Atoi(ipPort[1])
							port.ExternalIP = actualItems[2]
							if len(actualItems) > 4 {
								port.State = actualItems[3]
								port.PID, _ = strconv.Atoi(actualItems[4])
							} else {
								port.PID, _ = strconv.Atoi(actualItems[3])
							}

							ports = append(ports, port)
						}
					}
				}
			initAdapter:
				netAdapter := &NetAdapterInfo{}
				netAdapter.Name = interf.Name
				netAdapter.MacAddr = interf.HardwareAddr.String()
				netAdapter.IP = addr.String()
				netAdapter.IsUp = up
				netAdapter.IsIpv4 = isIpv4
				netAdapter.Ports = ports
				netAdapters = append(netAdapters, netAdapter)
				// }
			}
		}
	}
	sort.Slice(netAdapters, func(i, j int) bool {
		return netAdapters[i].IsUp && !netAdapters[j].IsUp
	})
	return netAdapters
}
