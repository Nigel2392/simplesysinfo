package simplesysinfo

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"

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
	Hostname string               `json:"hostname"`
	Platform string               `json:"platform"`
	CPU      CPUInfo              `json:"cpu"`
	RAM      RAMInfo              `json:"ram"`
	Disk     DiskInfo             `json:"disk"`
	Procs    map[int]*ProcessInfo `json:"procs"`
	MacAddr  string               `json:"macaddr"`
}

type ProcessInfo struct {
	Pid        int32  `json:"pid"`
	Name       string `json:"name"`
	Executable string `json:"executable"`
	Username   string `json:"username"`
}

func (p *ProcessInfo) String() string {
	return p.Name
}

// CPUInfo saves the CPU information
type CPUInfo struct {
	Threads      int32   `json:"threads"`
	ClockSpeed   float64 `json:"clockspeed"`
	CurrentUsage float32 `json:"currentusage"`
	Name         string  `json:"name"`
}

func (c *CPUInfo) String() string {
	return c.Name
}

// RAMInfo saves the RAM information
type RAMInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Swap  uint64 `json:"swap"`
}

func (r *RAMInfo) GetUsedPercentage() float64 {
	return float64(r.Used) / float64(r.Total) * 100
}

func (r *RAMInfo) String() string {
	return "<Memory> Total: " + ByteToGB(r.Total) + " Used: " + strconv.FormatFloat(r.GetUsedPercentage(), 'f', 2, 64) + "%"
}

// DiskInfo saves the Disk information
type DiskInfo struct {
	SysID string `json:"sysid"`
	Path  string `json:"path"`
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

func (d *DiskInfo) GetUsedPercentage() float64 {
	return float64(d.Used) / float64(d.Total) * 100
}

func (d *DiskInfo) String() string {
	return "<Disk> Total: " + ByteToGB(d.Total) + " Used: " + strconv.FormatFloat(d.GetUsedPercentage(), 'f', 2, 64) + "%"
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
			CurrentUsage: GetCPUUsage(50),
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

func ByteToGB(b uint64) string {
	return strconv.FormatFloat(float64(b)/1024/1024/1024, 'f', 2, 64) + " GB"
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

// Get cpu usage in percentage
func GetCPUUsage(ms int) float32 {
	percent, err := cpu.Percent(0, true)
	if err != nil {
		return 0
	}
	total := 0.0
	for _, p := range percent {
		total += p
	}
	return float32(total / float64(len(percent)))
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
