import { http } from '@/utils'

export default {
  // 获取进程列表 (支持排序、筛选、搜索)
  list: (params: {
    page: number
    limit: number
    sort_by?: string
    sort_dir?: string
    status?: string
    search?: string
    username?: string
  }) => http.Get('/process', { params }),
  
  // 获取进程树
  tree: () => http.Get('/process/tree'),
  
  // 获取进程详情
  detail: (pid: number) => http.Get('/process/detail', { params: { pid } }),
  
  // 杀死进程
  kill: (pid: number) => http.Post('/process/kill', { pid }),
  
  // 发送信号给进程
  signal: (pid: number, signal: string) => http.Post('/process/signal', { pid, signal })
}
