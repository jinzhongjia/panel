package service

import (
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/libtnb/chix"
	"github.com/shirou/gopsutil/process"

	"github.com/tnborg/panel/internal/http/request"
	"github.com/tnborg/panel/pkg/types"
)

type ProcessService struct {
	cache      map[int32]*types.ProcessData
	cacheMutex sync.RWMutex
	lastUpdate time.Time
	cacheTTL   time.Duration
}

func NewProcessService() *ProcessService {
	return &ProcessService{
		cache:    make(map[int32]*types.ProcessData),
		cacheTTL: 5 * time.Second,
	}
}
func (s *ProcessService) List(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ProcessList](r)
	if err != nil {
		req = &request.ProcessList{
			Page:  1,
			Limit: 50,
		}
	}

	data, err := s.getProcessList(req)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	if req.ShowTree {
		tree := s.buildProcessTree(data)
		Success(w, chix.M{
			"total": len(data),
			"items": tree,
			"tree":  true,
		})
		return
	}

	paged, total := s.paginateProcesses(data, req.Page, req.Limit)

	Success(w, chix.M{
		"total": total,
		"items": paged,
	})
}

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

	s.invalidateCache()
	Success(w, nil)
}

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

	signal := s.parseSignal(req.Signal)
	if err = proc.SendSignal(signal); err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	s.invalidateCache()
	Success(w, nil)
}

func (s *ProcessService) Detail(w http.ResponseWriter, r *http.Request) {
	req, err := Bind[request.ProcessDetail](r)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "%v", err)
		return
	}

	proc, err := process.NewProcess(req.PID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	detail, err := s.getProcessDetail(proc)
	if err != nil {
		Error(w, http.StatusInternalServerError, "%v", err)
		return
	}

	Success(w, detail)
}

// processProcess 处理进程数据
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
		data.MemPercent = memPercent
	}

	if ioStat, err := proc.IOCounters(); err == nil {
		data.DiskWrite = ioStat.WriteBytes
		data.DiskRead = ioStat.ReadBytes
	}

	data.Nets, _ = proc.NetIOCounters(false)
	data.Connections, _ = proc.Connections()
	data.CmdLine, _ = proc.Cmdline()
	data.OpenFiles, _ = proc.OpenFiles()
	data.Envs, _ = proc.Environ()
	data.OpenFiles = slices.Compact(data.OpenFiles)
	data.Envs = slices.Compact(data.Envs)

	// 新增字段
	if children, err := proc.Children(); err == nil {
		childPids := make([]int32, len(children))
		for i, child := range children {
			childPids[i] = child.Pid
		}
		data.Children = childPids
	}
	data.Exe, _ = proc.Exe()
	data.Cwd, _ = proc.Cwd()
	data.Nice, _ = proc.Nice()
	data.IONice, _ = proc.IOnice()
	data.IsRunning, _ = proc.IsRunning()
	data.Terminal, _ = proc.Terminal()
	data.Gids, _ = proc.Gids()
	data.Uids, _ = proc.Uids()

	return data
}

// getProcessList 获取进程列表
func (s *ProcessService) getProcessList(req *request.ProcessList) ([]types.ProcessData, error) {
	s.cacheMutex.RLock()
	if time.Since(s.lastUpdate) < s.cacheTTL && len(s.cache) > 0 {
		data := make([]types.ProcessData, 0, len(s.cache))
		for _, proc := range s.cache {
			data = append(data, *proc)
		}
		s.cacheMutex.RUnlock()
		return s.filterAndSortProcesses(data, req), nil
	}
	s.cacheMutex.RUnlock()

	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	data := make([]types.ProcessData, 0, len(processes))
	newCache := make(map[int32]*types.ProcessData)

	// 使用并发处理提高性能
	type result struct {
		data types.ProcessData
		err  error
	}

	resultChan := make(chan result, len(processes))
	semaphore := make(chan struct{}, 50) // 限制并发数

	for _, proc := range processes {
		go func(p *process.Process) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			processData := s.processProcess(p)
			resultChan <- result{data: processData}
		}(proc)
	}

	for i := 0; i < len(processes); i++ {
		res := <-resultChan
		if res.err == nil {
			data = append(data, res.data)
			newCache[res.data.PID] = &res.data
		}
	}

	s.cacheMutex.Lock()
	s.cache = newCache
	s.lastUpdate = time.Now()
	s.cacheMutex.Unlock()

	return s.filterAndSortProcesses(data, req), nil
}

// filterAndSortProcesses 过滤和排序进程
func (s *ProcessService) filterAndSortProcesses(data []types.ProcessData, req *request.ProcessList) []types.ProcessData {
	// 状态过滤
	if req.Status != "" {
		filtered := make([]types.ProcessData, 0)
		for _, proc := range data {
			if strings.EqualFold(proc.Status, req.Status) {
				filtered = append(filtered, proc)
			}
		}
		data = filtered
	}

	// 搜索过滤
	if req.Search != "" {
		search := strings.ToLower(req.Search)
		filtered := make([]types.ProcessData, 0)
		for _, proc := range data {
			if strings.Contains(strings.ToLower(proc.Name), search) ||
				strings.Contains(strings.ToLower(proc.CmdLine), search) ||
				strings.Contains(strings.ToLower(proc.Username), search) ||
				strings.Contains(strconv.Itoa(int(proc.PID)), search) {
				filtered = append(filtered, proc)
			}
		}
		data = filtered
	}

	// 排序
	sort.Slice(data, func(i, j int) bool {
		switch req.SortBy {
		case "name":
			if req.SortDesc {
				return data[i].Name > data[j].Name
			}
			return data[i].Name < data[j].Name
		case "cpu":
			if req.SortDesc {
				return data[i].CPU > data[j].CPU
			}
			return data[i].CPU < data[j].CPU
		case "memory":
			if req.SortDesc {
				return data[i].RSS > data[j].RSS
			}
			return data[i].RSS < data[j].RSS
		case "start_time":
			if req.SortDesc {
				return data[i].StartTime > data[j].StartTime
			}
			return data[i].StartTime < data[j].StartTime
		default: // pid
			if req.SortDesc {
				return data[i].PID > data[j].PID
			}
			return data[i].PID < data[j].PID
		}
	})

	return data
}

// buildProcessTree 构建进程树
func (s *ProcessService) buildProcessTree(data []types.ProcessData) []*types.ProcessTree {
	processMap := make(map[int32]*types.ProcessTree)
	var roots []*types.ProcessTree

	// 创建所有节点
	for _, proc := range data {
		tree := &types.ProcessTree{
			ProcessData: proc,
			Children:    make([]*types.ProcessTree, 0),
		}
		processMap[proc.PID] = tree
	}

	// 构建树结构
	for _, tree := range processMap {
		if tree.PPID == 0 || processMap[tree.PPID] == nil {
			roots = append(roots, tree)
		} else {
			parent := processMap[tree.PPID]
			parent.Children = append(parent.Children, tree)
		}
	}

	return roots
}

// paginateProcesses 分页处理
func (s *ProcessService) paginateProcesses(data []types.ProcessData, page, limit int) ([]types.ProcessData, int) {
	total := len(data)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	start := (page - 1) * limit
	if start >= total {
		return []types.ProcessData{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return data[start:end], total
}

// getProcessDetail 获取进程详情
func (s *ProcessService) getProcessDetail(proc *process.Process) (*types.ProcessDetail, error) {
	detail := &types.ProcessDetail{
		ProcessData: s.processProcess(proc),
	}

	if cpuTimes, err := proc.Times(); err == nil {
		detail.CPUTimes = cpuTimes
	}

	if memInfo, err := proc.MemoryInfo(); err == nil {
		detail.MemoryInfo = memInfo
	}

	if memMaps, err := proc.MemoryMaps(false); err == nil {
		detail.MemoryMaps = *memMaps
	}

	if pageFaults, err := proc.PageFaults(); err == nil {
		detail.PageFaults = pageFaults
	}

	if rlimit, err := proc.RlimitUsage(false); err == nil {
		detail.RlimitUsage = rlimit
	}

	return detail, nil
}

// parseSignal 解析信号
func (s *ProcessService) parseSignal(signal string) syscall.Signal {
	switch signal {
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGINT":
		return syscall.SIGINT
	case "SIGQUIT":
		return syscall.SIGQUIT
	case "SIGHUP":
		return syscall.SIGHUP
	case "SIGSTOP":
		return syscall.Signal(19) // SIGSTOP
	case "SIGCONT":
		return syscall.Signal(18) // SIGCONT
	case "SIGUSR1":
		return syscall.Signal(10) // SIGUSR1
	case "SIGUSR2":
		return syscall.Signal(12) // SIGUSR2
	default:
		return syscall.SIGTERM
	}
}

// invalidateCache 清除缓存
func (s *ProcessService) invalidateCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.cache = make(map[int32]*types.ProcessData)
	s.lastUpdate = time.Time{}
}
