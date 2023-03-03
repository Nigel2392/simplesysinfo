package simplesysinfo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type includedItem int

const (
	INC_HOSTNAME includedItem = iota
	INC_PLATFORM
	INC_CPU
	INC_MEM
	INC_DISK
	INC_MACADDR
	INC_PROCS
	INC_NETADAPTERS
	last_iota
)

var IncludeAll = getIncludeAll()

var VERBOSE = false

// SysInfo saves the basic system information
type SysInfo struct {
	Hostname    string               `json:"hostname"`
	Platform    string               `json:"platform"`
	CPU         *CPUInfo             `json:"cpu"`
	RAM         *RAMInfo             `json:"ram"`
	Disk        *DiskInfo            `json:"disk"`
	Procs       map[int]*ProcessInfo `json:"procs"`
	MainMacAddr string               `json:"macaddr"`
	NetAdapters NetAdapters          `json:"netadapters"`
}

func (s *SysInfo) String() string {
	var b strings.Builder
	b.WriteString("Hostname:\n\t")
	b.WriteString(s.Hostname)
	b.WriteString("\nPlatform:\n\t")
	b.WriteString(s.Platform)

	if s.CPU != nil {
		b.WriteString("\nCPU:\n")
		writeToBuf(&b, "Name", s.CPU.Name)
		writeToBufIfVerbose(&b, "Threads", s.CPU.Threads)
		writeToBufIfVerbose(&b, "Current Usage", fmt.Sprintf("%.2f", s.CPU.CurrentUsage))
		writeToBufIfVerbose(&b, "Clock Speed", s.CPU.ClockSpeed)
	}
	if s.RAM != nil {
		b.WriteString("RAM:\n")
		writeToBuf(&b, "Percentage Used", fmt.Sprintf("%.2f", s.RAM.GetUsedPercentage()))
		writeToBufIfVerbose(&b, "Total", ByteToGB(s.RAM.Total))
		writeToBufIfVerbose(&b, "Used", ByteToGB(s.RAM.Used))
		writeToBufIfVerbose(&b, "Free", ByteToGB(s.RAM.Free))
		writeToBufIfVerbose(&b, "Swap", ByteToGB(s.RAM.Swap))
		if VERBOSE {
			b.WriteString("\n")
		}
	}
	if s.Disk != nil {
		b.WriteString("Disk:\n")
		if VERBOSE && s.Disk.SysID != "" && s.Disk.Path == "" {
			writeToBuf(&b, "SysID", s.Disk.SysID)
		} else if VERBOSE && s.Disk.SysID != "" && s.Disk.Path != "" {
			writeToBuf(&b, "SysID", fmt.Sprintf("%s (%s)", s.Disk.SysID, s.Disk.Path))
		} else if VERBOSE && s.Disk.SysID == "" && s.Disk.Path != "" {
			writeToBuf(&b, "Path", s.Disk.Path)
		}
		writeToBufIfVerbose(&b, "Used Percentage", s.Disk.GetUsedPercentage())
		writeToBufIfVerbose(&b, "Total", ByteToGB(s.Disk.Total))
		writeToBuf(&b, "Used", ByteToGB(s.Disk.Used))
		writeToBufIfVerbose(&b, "Free", ByteToGB(s.Disk.Free))
	}
	if s.Procs != nil {
		if VERBOSE {
			b.WriteString("\nProcesses:\n")
			for _, p := range s.Procs {
				writeToBuf(&b, fmt.Sprintf("(%d) %s", p.Pid, p.Name), p.Executable)
			}
		} else {
			b.WriteString(fmt.Sprintf("\nProcesses: %d", len(s.Procs)))
		}
		b.WriteString("\n")
	}
	if s.NetAdapters != nil {
		b.WriteString(fmt.Sprintf("\nNetwork Adapters: %d\n", len(s.NetAdapters)))
		for _, netAdapter := range s.NetAdapters {
			writeToBuf(&b, "Name", netAdapter.Name)
			writeToBuf(&b, "IsUp", netAdapter.IsUp)
			writeToBuf(&b, "IsIpv4", netAdapter.IsIpv4)
			writeToBufIfVerbose(&b, "IP", netAdapter.IP)
			writeToBufIfVerbose(&b, "MacAddr", netAdapter.MacAddr)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func GetSysInfo(include ...includedItem) *SysInfo {
	hostStat, _ := host.Info()
	info := new(SysInfo)
	if len(include) == 0 {
		include = IncludeAll
	}
	if Contains(include, INC_HOSTNAME) {
		info.Hostname = strings.TrimSpace(hostStat.Hostname)
	}
	if Contains(include, INC_PLATFORM) {
		info.Platform = strings.TrimSpace(hostStat.Platform)
	}
	if Contains(include, INC_CPU) {
		cpuStat, _ := cpu.Info()
		info.CPU = &CPUInfo{
			Threads:      cpuStat[0].Cores,
			ClockSpeed:   cpuStat[0].Mhz,
			CurrentUsage: GetCPUUsage(50),
			Name:         strings.TrimSpace(cpuStat[0].ModelName),
		}
	}
	if Contains(include, INC_MEM) {
		vmStat, _ := mem.VirtualMemory()
		info.RAM = &RAMInfo{
			Total: vmStat.Total, // 1024 / 1024, // MB
			Used:  vmStat.Used,  // 1024 / 1024,  // MB
			Free:  vmStat.Free,  // 1024 / 1024,  // MB
			Swap:  vmStat.SwapTotal,
		}
	}
	if Contains(include, INC_DISK) {
		diskStat, _ := disk.Usage("\\") // If you're in Unix change this "\\" for "/"
		info.Disk = &DiskInfo{
			Path:  diskStat.Path,
			Total: diskStat.Total, // 1024 / 1024 / 1024, // GB
			Used:  diskStat.Used,  // 1024 / 1024 / 1024,  // GB
			Free:  diskStat.Free,  // 1024 / 1024 / 1024,  // GB
		}
	}
	if Contains(include, INC_MACADDR) {
		info.MainMacAddr, _ = GetMACAddr()
	}
	if Contains(include, INC_PROCS) {
		procs, _ := GetProcs()
		info.Procs = procs
	}
	if Contains(include, INC_NETADAPTERS) {
		info.NetAdapters = GetNetAdapters()
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
