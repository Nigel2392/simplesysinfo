package simplesysinfo

import (
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
)

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
	return "Memory Total: " + ByteToGB(r.Total) + " Used: " + strconv.FormatFloat(r.GetUsedPercentage(), 'f', 2, 64) + "%"
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
	return "Disk Total: " + ByteToGB(d.Total) + " Used: " + strconv.FormatFloat(d.GetUsedPercentage(), 'f', 2, 64) + "%"
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
