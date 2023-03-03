package simplesysinfo

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
)

type ProcessInfo struct {
	Pid        int32  `json:"pid"`
	Name       string `json:"name"`
	Executable string `json:"executable"`
	Username   string `json:"username"`
	Memory     *process.MemoryInfoStat
}

func (p *ProcessInfo) String() string {
	return p.Name
}

type CPU struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physicalId"`
	CoreID     string   `json:"coreId"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cacheSize"`
	Flags      []string `json:"flags"`
	Microcode  string   `json:"microcode"`
}

// CPUInfo saves the CPU information
type CPUInfo struct {
	CPUs         []*CPU  `json:"cpus"`
	CurrentUsage float32 `json:"currentusage"`
}

func (c *CPUInfo) String() string {
	return strings.Join(func() []string {
		var s []string
		for _, cpu := range c.CPUs {
			s = append(s, cpu.ModelName)
		}
		return s
	}(), ", ")
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

func GetProcs() ([]*ProcessInfo, error) {
	procs := make([]*ProcessInfo, 0)
	os_procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(os_procs))
	for _, p := range os_procs {
		go func(p *process.Process) {
			defer wg.Done()
			proc := &ProcessInfo{}
			proc.Pid = p.Pid
			var nameChan = make(chan string, 1)
			var exeChan = make(chan string, 1)
			var usernameChan = make(chan string, 1)
			var memoryInfoStatChan = make(chan *process.MemoryInfoStat, 1)
			go func() {
				name, _ := p.Name()
				nameChan <- name
			}()
			go func() {
				exe, _ := p.Exe()
				exeChan <- exe
			}()
			go func() {
				username, _ := p.Username()
				usernameChan <- username
			}()
			go func() {
				memoryInfoStat, _ := p.MemoryInfo()
				memoryInfoStatChan <- memoryInfoStat
			}()

			for i := 0; i < 4; i++ {
				select {
				case name := <-nameChan:
					proc.Name = name
				case exe := <-exeChan:
					proc.Executable = exe
				case username := <-usernameChan:
					proc.Username = username
				case memoryInfoStat := <-memoryInfoStatChan:
					proc.Memory = memoryInfoStat
				}
			}

			mu.Lock()
			procs = append(procs, proc)
			mu.Unlock()
		}(p)
	}
	wg.Wait()
	return procs, nil
}

// Get cpu usage in percentage
func GetCPUUsage(ms int) float32 {
	percent, err := cpu.Percent(time.Millisecond*time.Duration(ms), true)
	if err != nil {
		return 0
	}
	total := 0.0
	for _, p := range percent {
		total += p
	}
	return float32(total / float64(len(percent)))
}
