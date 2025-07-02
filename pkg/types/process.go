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

	// 新增字段，用于详细进程信息
	Exe         string   `json:"exe"`          // 可执行文件路径
	Cwd         string   `json:"cwd"`          // 工作目录
	CreateTime  int64    `json:"create_time"`  // 创建时间戳
	Children    []int32  `json:"children"`     // 子进程PID列表
	CPUTimes    *cpu.TimesStat `json:"cpu_times"`    // CPU时间统计
	MemoryPercent float32 `json:"memory_percent"` // 内存使用百分比
	FDs         int32    `json:"fds"`          // 文件描述符数量
	GIDs        []int32  `json:"gids"`         // 组ID
	UIDs        []int32  `json:"uids"`         // 用户ID
	Terminal    string   `json:"terminal"`     // 终端
	Nice        int32    `json:"nice"`         // 优先级
}

// ProcessListRequest 进程列表请求参数
type ProcessListRequest struct {
	Page     int    `form:"page" json:"page"`
	Limit    int    `form:"limit" json:"limit"`
	SortBy   string `form:"sort_by" json:"sort_by"`     // pid, cpu, memory, start_time, name
	SortDir  string `form:"sort_dir" json:"sort_dir"`   // asc, desc
	Status   string `form:"status" json:"status"`       // 状态筛选
	Search   string `form:"search" json:"search"`       // 搜索关键词
	Username string `form:"username" json:"username"`   // 用户筛选
}

// ProcessTreeNode 进程树节点
type ProcessTreeNode struct {
	ProcessData
	Children []*ProcessTreeNode `json:"children"`
	Level    int                `json:"level"` // 树层级
}

// ProcessDetailData 进程详细信息
type ProcessDetailData struct {
	ProcessData
	CommandLine []string                `json:"command_line"` // 完整命令行参数
	Parent      *ProcessData            `json:"parent"`       // 父进程信息
	ChildrenDetail []ProcessData        `json:"children_detail"` // 子进程详细信息
}
