# 进程管理API测试

## 测试前后端集成

### 1. 启动后端服务
```bash
cd C:\Users\jin\code\panel
go run cmd/web/main.go
```

### 2. 测试API接口

#### 获取进程列表（基础）
```bash
curl "http://localhost:8080/api/process?page=1&limit=10"
```

#### 获取进程列表（带搜索）
```bash
curl "http://localhost:8080/api/process?page=1&limit=10&search=nginx"
```

#### 获取进程列表（带状态过滤）
```bash
curl "http://localhost:8080/api/process?page=1&limit=10&status=running"
```

#### 获取进程列表（带排序）
```bash
curl "http://localhost:8080/api/process?page=1&limit=10&sort_by=cpu&sort_desc=true"
```

#### 获取进程树视图
```bash
curl "http://localhost:8080/api/process?page=1&limit=10&show_tree=true"
```

#### 发送信号到进程
```bash
curl -X POST "http://localhost:8080/api/process/signal" \
  -H "Content-Type: application/json" \
  -d '{"pid": 1234, "signal": "SIGTERM"}'
```

#### 获取进程详情
```bash
curl "http://localhost:8080/api/process/1234/detail"
```

#### 终止进程（兼容接口）
```bash
curl -X POST "http://localhost:8080/api/process/kill" \
  -H "Content-Type: application/json" \
  -d '{"pid": 1234}'
```

### 3. 前端功能测试

1. 打开浏览器访问 `http://localhost:8080`
2. 登录后导航到 "任务" -> "系统进程"
3. 测试以下功能：
   - 搜索进程
   - 按状态过滤
   - 排序功能
   - 树状视图切换
   - 查看进程详情
   - 发送信号
   - 终止进程

### 4. 新增功能验证

#### 多信号支持
- 测试发送不同类型的信号（SIGTERM, SIGKILL, SIGSTOP, SIGCONT等）
- 验证信号发送后的进程状态变化

#### 高级过滤和排序
- 测试按不同字段排序（PID, 名称, CPU, 内存, 启动时间）
- 测试升序/降序排列
- 测试状态过滤功能

#### 进程详情页
- 验证基础信息显示
- 验证资源使用信息
- 验证系统信息
- 验证命令行和环境变量显示

#### 性能优化
- 观察页面加载速度
- 测试大量进程时的响应性能
- 验证缓存机制是否生效

### 5. 向后兼容性测试

确保原有的API接口仍然正常工作：
- `/api/process` 基础列表接口
- `/api/process/kill` 终止进程接口

所有新功能都通过可选参数提供，不影响现有功能。