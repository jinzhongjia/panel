package types

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type ProcessData struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	PPID       int32   `json:"ppid"`
	Username   string  `json:"username"`
	Status     string  `json:"status"`
	Background bool    `json:"background"`
	StartTime  string  `json:"start_time"`
	NumThreads int32   `json:"num_threads"`
	CPU        float64 `json:"cpu"`

	DiskRead  uint64 `json:"disk_read"`
	DiskWrite uint64 `json:"disk_write"`

	CmdLine string `json:"cmd_line"`

	RSS    uint64 `json:"rss"`
	VMS    uint64 `json:"vms"`
	HWM    uint64 `json:"hwm"`
	Data   uint64 `json:"data"`
	Stack  uint64 `json:"stack"`
	Locked uint64 `json:"locked"`
	Swap   uint64 `json:"swap"`

	Envs []string `json:"envs"`

	OpenFiles   []process.OpenFilesStat `json:"open_files"`
	Connections []net.ConnectionStat    `json:"connections"`
	Nets        []net.IOCountersStat    `json:"nets"`

	// 新增字段
	Children   []int32 `json:"children"`    // 子进程PID列表
	Exe        string  `json:"exe"`         // 可执行文件路径
	Cwd        string  `json:"cwd"`         // 当前工作目录
	Nice       int32   `json:"nice"`        // 进程优先级
	IONice     int32   `json:"ionice"`      // IO优先级
	MemPercent float32 `json:"mem_percent"` // 内存使用百分比
	IsRunning  bool    `json:"is_running"`  // 是否正在运行
	Terminal   string  `json:"terminal"`    // 终端
	Gids       []int32 `json:"gids"`        // 组ID列表
	Uids       []int32 `json:"uids"`        // 用户ID列表
}

type ProcessTree struct {
	ProcessData
	Children []*ProcessTree `json:"children"`
}

type ProcessDetail struct {
	ProcessData
	CPUTimes    *cpu.TimesStat           `json:"cpu_times"`
	MemoryInfo  *process.MemoryInfoStat  `json:"memory_info"`
	MemoryMaps  []process.MemoryMapsStat `json:"memory_maps"`
	PageFaults  *process.PageFaultsStat  `json:"page_faults"`
	RlimitUsage []process.RlimitStat     `json:"rlimit_usage"`
}
