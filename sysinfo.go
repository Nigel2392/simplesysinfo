package simplesysinfo

import (
	"encoding/json"
	"math"
	"net"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

const (
	INC_HOSTNAME = iota
	INC_PLATFORM
	INC_CPU
	INC_MEM
	INC_DISK
	INC_MACADDR
	INC_PROCS
	last_iota
)

func GetIncludeAll() []int {
	var includes []int
	for i := 0; i < last_iota; i++ {
		includes = append(includes, i)
	}
	return includes
}

var IncludeAll = GetIncludeAll()

// SysInfo saves the basic system information
type SysInfo struct {
	Hostname string               `json:"hostname,omitempty"`
	Platform string               `json:"platform,omitempty"`
	CPU      CPUInfo              `json:"cpu,omitempty"`
	RAM      RAMInfo              `json:"ram,omitempty"`
	Disk     DiskInfo             `json:"disk,omitempty"`
	Procs    map[int]*ProcessInfo `json:"procs,omitempty"`
	MacAddr  string               `json:"macaddr,omitempty"`
}

type ProcessInfo struct {
	Pid        int32  `json:"pid,omitempty"`
	Name       string `json:"name,omitempty"`
	Executable string `json:"executable,omitempty"`
	Username   string `json:"username,omitempty"`
}

// CPUInfo saves the CPU information
type CPUInfo struct {
	Threads      int32   `json:"threads,omitempty"`
	ClockSpeed   float64 `json:"clockspeed,omitempty"`
	CurrentUsage int     `json:"currentusage,omitempty"`
	Name         string  `json:"name,omitempty"`
}

// RAMInfo saves the RAM information
type RAMInfo struct {
	Total uint64 `json:"total,omitempty"`
	Used  uint64 `json:"used,omitempty"`
	Free  uint64 `json:"free,omitempty"`
	Swap  uint64 `json:"swap,omitempty"`
}

// DiskInfo saves the Disk information
type DiskInfo struct {
	SysID string `json:"sysid,omitempty"`
	Path  string `json:"path,omitempty"`
	Total uint64 `json:"total,omitempty"`
	Used  uint64 `json:"used,omitempty"`
	Free  uint64 `json:"free,omitempty"`
}

func GetSysInfo(include []int) *SysInfo {
	hostStat, _ := host.Info()
	info := new(SysInfo)
	if ContainsInt(include, INC_HOSTNAME) {
		info.Hostname = strings.TrimSpace(hostStat.Hostname)
	}
	if ContainsInt(include, INC_PLATFORM) {
		info.Platform = strings.TrimSpace(hostStat.Platform)
	}
	if ContainsInt(include, INC_CPU) {
		cpuStat, _ := cpu.Info()
		info.CPU = CPUInfo{
			Threads:      cpuStat[0].Cores,
			ClockSpeed:   cpuStat[0].Mhz,
			CurrentUsage: GetCPUUsage(),
			Name:         strings.TrimSpace(cpuStat[0].ModelName),
		}
	}
	if ContainsInt(include, INC_MEM) {
		vmStat, _ := mem.VirtualMemory()
		info.RAM = RAMInfo{
			Total: vmStat.Total, // 1024 / 1024, // MB
			Used:  vmStat.Used,  // 1024 / 1024,  // MB
			Free:  vmStat.Free,  // 1024 / 1024,  // MB
			Swap:  vmStat.SwapTotal,
		}
	}
	if ContainsInt(include, INC_DISK) {
		diskStat, _ := disk.Usage("\\") // If you're in Unix change this "\\" for "/"
		info.Disk = DiskInfo{
			Path:  diskStat.Path,
			Total: diskStat.Total, // 1024 / 1024 / 1024, // GB
			Used:  diskStat.Used,  // 1024 / 1024 / 1024,  // GB
			Free:  diskStat.Free,  // 1024 / 1024 / 1024,  // GB
		}
	}
	if ContainsInt(include, INC_MACADDR) {
		info.MacAddr, _ = GetMACAddr()
	}
	if ContainsInt(include, INC_PROCS) {
		procs, _ := GetProcs()
		info.Procs = procs
	}
	return info
}

func (s *SysInfo) ToJSON() []byte {
	json, _ := json.Marshal(s)
	return json
}

func (s *SysInfo) FromJson(jdata []byte) *SysInfo {
	json.Unmarshal(jdata, s)
	return s
}

func GetMACAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	var currentIP, currentNetworkHardwareName string
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		// = GET LOCAL IP ADDRESS
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentIP = ipnet.IP.String()
			}
		}
	}
	// get all the system's or local machine's network interfaces
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				// only interested in the name with current IP address
				if strings.Contains(addr.String(), currentIP) {
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}
	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)
	if err != nil {
		return "", err
	}
	macAddress := netInterface.HardwareAddr
	// verify if the MAC address can be parsed properly
	hwAddr, err := net.ParseMAC(macAddress.String())
	if err != nil {
		return "", err
	}
	return hwAddr.String(), nil
}

func ContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func GetProcs() (map[int]*ProcessInfo, error) {
	procs := make(map[int]*ProcessInfo)
	os_procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range os_procs {
		proc := &ProcessInfo{}
		proc.Pid = p.Pid
		proc.Name, _ = p.Name()
		proc.Executable, _ = p.Exe()
		proc.Username, _ = p.Username()
		procs[int(p.Pid)] = proc
	}
	return procs, nil
}
func GetCPUUsage() int {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0
	}
	return int(math.Round(percent[0]))
}

//	// Client
//	if len(c.CONFIG.Included_info) > 0 {
//		b16data := quickproto.Base64Encoding(sysinfo.GetSysInfo(c.CONFIG.Included_info).ToJSON())
//		msg.AddHeader("sysinfo", string(b16data))
//	}
//	// Server
//	msg, err := quickproto.ReadConn(client.Conn, s.CONFIG, client.Key)
//	if err != nil {
//		return &quickproto.Message{}, err
//	}
//	if len(s.CONFIG.Included_info) > 0 {
//		var info sysinfo.SysInfo
//		var sysinfo_serialized []byte
//		var err error
//		// Get info from header.
//		sysinfo_enc := msg.Headers["sysinfo"][0]
//		// Decode the sysinfo from base32.
//		if sysinfo_serialized, err = quickproto.Base64Decoding([]byte(sysinfo_enc)); err != nil {
//			return msg, err
//		}
//		client.SysInfo = info.FromJson(sysinfo_serialized)
//		delete(msg.Headers, "sysinfo")
//	}
