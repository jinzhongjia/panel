package service

import (
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-rat/chix"
	"github.com/shirou/gopsutil/process"

	"github.com/tnb-labs/panel/internal/http/request"
	"github.com/tnb-labs/panel/pkg/types"
)

type ProcessService struct {
}

func NewProcessService() *ProcessService {
	return &ProcessService{}
}

// List 获取进程列表（支持排序、筛选、搜索）
func (s *ProcessService) List(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ProcessListRequest](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	processes, err := process.Processes()
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	data := make([]types.ProcessData, 0)
	for proc := range slices.Values(processes) {
		processData := s.processProcess(proc)
		
		// 状态筛选
		if req.Status != "" && processData.Status != req.Status {
			continue
		}
		
		// 用户筛选
		if req.Username != "" && processData.Username != req.Username {
			continue
		}
		
		// 搜索筛选
		if req.Search != "" {
			searchLower := strings.ToLower(req.Search)
			if !strings.Contains(strings.ToLower(processData.Name), searchLower) &&
				!strings.Contains(strings.ToLower(processData.CmdLine), searchLower) &&
				!strings.Contains(strconv.Itoa(int(processData.PID)), searchLower) {
				continue
			}
		}
		
		data = append(data, processData)
	}

	// 排序
	s.sortProcesses(data, req.SortBy, req.SortDir)

	// 分页
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	total := len(data)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	Success(w, chix.M{
		"total": total,
		"items": data[start:end],
	})
}

// Tree 获取进程树
func (s *ProcessService) Tree(w http.ResponseWriter, r *http.Request) {
	processes, err := process.Processes()
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	// 构建进程映射
	processMap := make(map[int32]*types.ProcessTreeNode)
	for proc := range slices.Values(processes) {
		processData := s.processProcess(proc)
		node := &types.ProcessTreeNode{
			ProcessData: processData,
			Children:    make([]*types.ProcessTreeNode, 0),
			Level:       0,
		}
		processMap[processData.PID] = node
	}

	// 构建树结构
	rootNodes := make([]*types.ProcessTreeNode, 0)
	for _, node := range processMap {
		if parent, exists := processMap[node.PPID]; exists && node.PID != node.PPID {
			parent.Children = append(parent.Children, node)
			node.Level = parent.Level + 1
		} else {
			rootNodes = append(rootNodes, node)
		}
	}

	// 排序根节点
	sort.Slice(rootNodes, func(i, j int) bool {
		return rootNodes[i].PID < rootNodes[j].PID
	})

	Success(w, rootNodes)
}

// Detail 获取进程详情
func (s *ProcessService) Detail(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ProcessDetailRequest](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	proc, err := process.NewProcess(req.PID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	processData := s.processProcess(proc)
	
	detail := types.ProcessDetailData{
		ProcessData: processData,
	}

	// 获取完整命令行参数
	if cmdlineSlice, err := proc.CmdlineSlice(); err == nil {
		detail.CommandLine = cmdlineSlice
	}

	// 获取父进程信息
	if ppid, err := proc.Ppid(); err == nil && ppid > 0 {
		if parentProc, err := process.NewProcess(ppid); err == nil {
			parentData := s.processProcess(parentProc)
			detail.Parent = &parentData
		}
	}

	// 获取子进程详细信息
	if children, err := proc.Children(); err == nil {
		detail.ChildrenDetail = make([]types.ProcessData, 0, len(children))
		for _, child := range children {
			childData := s.processProcess(child)
			detail.ChildrenDetail = append(detail.ChildrenDetail, childData)
		}
	}

	Success(w, detail)
}

// Kill 杀死进程
func (s *ProcessService) Kill(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ProcessKill](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	proc, err := process.NewProcess(req.PID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	if err = proc.Kill(); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

// Signal 发送信号给进程
func (s *ProcessService) Signal(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ProcessSignal](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	proc, err := process.NewProcess(req.PID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	// 将信号字符串转换为系统信号
	var sig syscall.Signal
	switch req.Signal {
	case "SIGTERM":
		sig = syscall.SIGTERM
	case "SIGKILL":
		sig = syscall.SIGKILL
	case "SIGINT":
		sig = syscall.SIGINT
	case "SIGHUP":
		sig = syscall.SIGHUP
	case "SIGUSR1":
		sig = syscall.SIGUSR1
	case "SIGUSR2":
		sig = syscall.SIGUSR2
	case "SIGQUIT":
		sig = syscall.SIGQUIT
	case "SIGSTOP":
		sig = syscall.SIGSTOP
	case "SIGCONT":
		sig = syscall.SIGCONT
	default:
		Error(w, http.StatusBadRequest, "不支持的信号类型: %s", req.Signal)
		return
	}

	if err = proc.SendSignal(sig); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, nil)
}

// processProcess 处理进程数据（增强版）
func (s *ProcessService) processProcess(proc *process.Process) types.ProcessData {
	data := types.ProcessData{
		PID: proc.Pid,
	}

	if name, err := proc.Name(); err == nil {
		data.Name = name
	} else {
		data.Name = "<UNKNOWN>"
	}

	if username, err := proc.Username(); err == nil {
		data.Username = username
	}
	data.PPID, _ = proc.Ppid()
	data.Status, _ = proc.Status()
	data.Background, _ = proc.Background()
	
	if ct, err := proc.CreateTime(); err == nil {
		data.CreateTime = ct
		data.StartTime = time.Unix(ct/1000, 0).Format(time.DateTime)
	}
	
	data.NumThreads, _ = proc.NumThreads()
	data.CPU, _ = proc.CPUPercent()
	
	if mem, err := proc.MemoryInfo(); err == nil {
		data.RSS = mem.RSS
		data.Data = mem.Data
		data.VMS = mem.VMS
		data.HWM = mem.HWM
		data.Stack = mem.Stack
		data.Locked = mem.Locked
		data.Swap = mem.Swap
	}

	if memPercent, err := proc.MemoryPercent(); err == nil {
		data.MemoryPercent = memPercent
	}

	if ioStat, err := proc.IOCounters(); err == nil {
		data.DiskWrite = ioStat.WriteBytes
		data.DiskRead = ioStat.ReadBytes
	}

	// 新增字段
	data.Exe, _ = proc.Exe()
	data.Cwd, _ = proc.Cwd()
	data.Terminal, _ = proc.Terminal()
	data.Nice, _ = proc.Nice()
	
	if children, err := proc.Children(); err == nil {
		data.Children = make([]int32, len(children))
		for i, child := range children {
			data.Children[i] = child.Pid
		}
	}
	
	data.CPUTimes, _ = proc.Times()
	
	if fds, err := proc.NumFDs(); err == nil {
		data.FDs = fds
	}
	
	if gids, err := proc.Gids(); err == nil {
		data.GIDs = gids
	}
	
	if uids, err := proc.Uids(); err == nil {
		data.UIDs = uids
	}

	data.Nets, _ = proc.NetIOCounters(false)
	data.Connections, _ = proc.Connections()
	data.CmdLine, _ = proc.Cmdline()
	data.OpenFiles, _ = proc.OpenFiles()
	data.Envs, _ = proc.Environ()
	data.OpenFiles = slices.Compact(data.OpenFiles)
	data.Envs = slices.Compact(data.Envs)

	return data
}

// sortProcesses 排序进程列表
func (s *ProcessService) sortProcesses(data []types.ProcessData, sortBy, sortDir string) {
	if sortBy == "" {
		sortBy = "pid"
	}
	if sortDir == "" {
		sortDir = "asc"
	}

	sort.Slice(data, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "pid":
			less = data[i].PID < data[j].PID
		case "cpu":
			less = data[i].CPU < data[j].CPU
		case "memory":
			less = data[i].RSS < data[j].RSS
		case "start_time":
			less = data[i].CreateTime < data[j].CreateTime
		case "name":
			less = strings.ToLower(data[i].Name) < strings.ToLower(data[j].Name)
		default:
			less = data[i].PID < data[j].PID
		}
		
		if sortDir == "desc" {
			return !less
		}
		return less
	})
}
