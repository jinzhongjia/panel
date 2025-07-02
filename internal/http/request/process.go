package request

// ProcessKill 杀死进程
type ProcessKill struct {
	PID int32 `json:"pid" validate:"required"`
}

// ProcessSignal 发送信号给进程
type ProcessSignal struct {
	PID    int32  `json:"pid" validate:"required"`
	Signal string `json:"signal" validate:"required|in:SIGTERM,SIGKILL,SIGINT,SIGHUP,SIGUSR1,SIGUSR2,SIGQUIT,SIGSTOP,SIGCONT"`
}

// ProcessListRequest 进程列表请求
type ProcessListRequest struct {
	Page     int    `form:"page" json:"page"`
	Limit    int    `form:"limit" json:"limit"`
	SortBy   string `form:"sort_by" json:"sort_by"`     // pid, cpu, memory, start_time, name
	SortDir  string `form:"sort_dir" json:"sort_dir"`   // asc, desc
	Status   string `form:"status" json:"status"`       // 状态筛选
	Search   string `form:"search" json:"search"`       // 搜索关键词
	Username string `form:"username" json:"username"`   // 用户筛选
}

// ProcessDetailRequest 进程详情请求
type ProcessDetailRequest struct {
	PID int32 `form:"pid" json:"pid" validate:"required"`
}
