package request

type ProcessKill struct {
	PID int32 `json:"pid" validate:"required"`
}

type ProcessSignal struct {
	PID    int32  `json:"pid" validate:"required"`
	Signal string `json:"signal" validate:"required,oneof=SIGTERM SIGKILL SIGINT SIGQUIT SIGSTOP SIGCONT SIGHUP SIGUSR1 SIGUSR2"`
}

type ProcessList struct {
	Page     int    `json:"page" validate:"min=1"`
	Limit    int    `json:"limit" validate:"min=1,max=1000"`
	SortBy   string `json:"sort_by" validate:"omitempty,oneof=pid name cpu memory start_time"`
	SortDesc bool   `json:"sort_desc"`
	Status   string `json:"status" validate:"omitempty,oneof=running sleeping stopped zombie idle"`
	Search   string `json:"search"`
	ShowTree bool   `json:"show_tree"`
}

type ProcessDetail struct {
	PID int32 `json:"pid" validate:"required"`
}
