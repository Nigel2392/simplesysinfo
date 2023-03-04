package simplesysinfo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/exp/slices"
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
	Hostname    string         `json:"hostname"`
	Platform    string         `json:"platform"`
	CPU         *CPUInfo       `json:"cpu"`
	RAM         *RAMInfo       `json:"ram"`
	Disk        *DiskInfo      `json:"disk"`
	Procs       []*ProcessInfo `json:"procs"`
	MainMacAddr string         `json:"macaddr"`
	NetAdapters NetAdapters    `json:"netadapters"`
}

func (s *SysInfo) String() string {
	var b strings.Builder
	b.WriteString("Hostname:\n\t")
	b.WriteString(s.Hostname)
	b.WriteString("\nPlatform:\n\t")
	b.WriteString(s.Platform)

	if s.CPU != nil {
		b.WriteString("\nCPU:\n")
		writeToBuf(&b, 1, "Current Usage", s.CPU.CurrentUsage)
		for _, c := range s.CPU.CPUs {
			writeToBuf(&b, 1, "ModelName", c.ModelName)
			writeToBuf(&b, 1, "Cores", c.Cores)
			writeToBufIfVerbose(&b, 1, "CPU", c.CPU)
			writeToBufIfVerbose(&b, 1, "VendorID", c.VendorID)
			writeToBufIfVerbose(&b, 1, "Family", c.Family)
			writeToBufIfVerbose(&b, 1, "Model", c.Model)
			writeToBufIfVerbose(&b, 1, "Stepping", c.Stepping)
			writeToBufIfVerbose(&b, 1, "PhysicalID", c.PhysicalID)
			writeToBufIfVerbose(&b, 1, "CoreID", c.CoreID)
			writeToBufIfVerbose(&b, 1, "Mhz", c.Mhz)
			writeToBufIfVerbose(&b, 1, "CacheSize", c.CacheSize)
			writeToBufIfVerbose(&b, 1, "Flags", c.Flags)
			writeToBufIfVerbose(&b, 1, "Microcode", c.Microcode)
		}
	}
	if s.RAM != nil {
		b.WriteString("RAM:\n")
		writeToBuf(&b, 1, "Percentage Used", fmt.Sprintf("%.2f", s.RAM.GetUsedPercentage()))
		writeToBufIfVerbose(&b, 1, "Total", ByteToGB(s.RAM.Total), "null")
		writeToBufIfVerbose(&b, 1, "Used", ByteToGB(s.RAM.Used), "null")
		writeToBufIfVerbose(&b, 1, "Free", ByteToGB(s.RAM.Free), "null")
		writeToBufIfVerbose(&b, 1, "Swap", ByteToGB(s.RAM.Swap), "null")
		if VERBOSE {
			b.WriteString("\n")
		}
	}
	if s.Disk != nil {
		b.WriteString("Disk:\n")
		if VERBOSE && s.Disk.SysID != "" && s.Disk.Path == "" {
			writeToBuf(&b, 1, "SysID", s.Disk.SysID)
		} else if VERBOSE && s.Disk.SysID != "" && s.Disk.Path != "" {
			writeToBuf(&b, 1, "SysID", fmt.Sprintf("%s (%s)", s.Disk.SysID, s.Disk.Path))
		} else if VERBOSE && s.Disk.SysID == "" && s.Disk.Path != "" {
			writeToBuf(&b, 1, "Path", s.Disk.Path)
		}
		writeToBufIfVerbose(&b, 1, "Used Percentage", s.Disk.GetUsedPercentage())
		writeToBufIfVerbose(&b, 1, "Total", ByteToGB(s.Disk.Total), "null")
		writeToBuf(&b, 1, "Used", ByteToGB(s.Disk.Used), "null")
		writeToBufIfVerbose(&b, 1, "Free", ByteToGB(s.Disk.Free), "null")
	}
	if s.Procs != nil {
		if VERBOSE {
			b.WriteString("\nProcesses:\n")
			for _, p := range s.Procs {
				writeToBuf(&b, 1, fmt.Sprintf("(%d) %s", p.Pid, p.Name), p.Executable, "")
			}
		} else {
			b.WriteString(fmt.Sprintf("\nProcesses: %d", len(s.Procs)))
		}
		b.WriteString("\n")
	}
	if s.NetAdapters != nil {
		b.WriteString(fmt.Sprintf("\nNetwork Adapters: %d\n", len(s.NetAdapters)))
		for _, netAdapter := range s.NetAdapters {
			writeToBuf(&b, 1, "Name", netAdapter.Name)
			writeToBuf(&b, 1, "IsUp", netAdapter.IsUp)
			writeToBuf(&b, 1, "IsIpv4", netAdapter.IsIpv4)
			writeToBufIfVerbose(&b, 1, "IP", netAdapter.IP, "error fetching IP")
			writeToBufIfVerbose(&b, 1, "MacAddr", netAdapter.MacAddr, "no MAC address found")
			if VERBOSE {
				for _, port := range netAdapter.Ports {
					writeToBuf(&b, 2, "Protocol", port.Protocol)
					writeToBuf(&b, 2, "Port", port.Port)
					writeToBuf(&b, 2, "State", port.State, "unknown")
					writeToBuf(&b, 2, "PID", port.PID)
					b.WriteString("\n")
				}
			} else {
				b.WriteString("Ports:" + strconv.Itoa(len(netAdapter.Ports)))
			}
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
			CurrentUsage: GetCPUUsage(50),
		}
		for _, cpu := range cpuStat {
			info.CPU.CPUs = append(info.CPU.CPUs, &CPU{
				CPU:        cpu.CPU,
				VendorID:   cpu.VendorID,
				Family:     cpu.Family,
				Model:      cpu.Model,
				Stepping:   cpu.Stepping,
				PhysicalID: cpu.PhysicalID,
				CoreID:     cpu.CoreID,
				Cores:      cpu.Cores,
				ModelName:  cpu.ModelName,
				Mhz:        cpu.Mhz,
				CacheSize:  cpu.CacheSize,
				Flags:      cpu.Flags,
				Microcode:  cpu.Microcode,
			})
		}
		slices.SortFunc(info.CPU.CPUs, func(i, j *CPU) bool {
			return i.Cores < j.Cores
		})
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
		slices.SortFunc(info.Procs, func(i, j *ProcessInfo) bool {
			return i.Executable > j.Executable
		})
	}
	if Contains(include, INC_NETADAPTERS) {
		info.NetAdapters = GetNetAdapters()
	}
	return info
}

func (s *SysInfo) JSON() []byte {
	json, _ := json.MarshalIndent(s, "", "  ")
	return json
}

func (s *SysInfo) UnJSON(jdata []byte) *SysInfo {
	json.Unmarshal(jdata, s)
	return s
}
