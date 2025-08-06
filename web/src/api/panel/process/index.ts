import { http } from '@/utils'

export interface ProcessListParams {
  page?: number
  limit?: number
  sort_by?: string
  sort_desc?: boolean
  status?: string
  search?: string
  show_tree?: boolean
}

export interface ProcessSignalParams {
  pid: number
  signal: string
}

export default {
  // 获取进程列表
  list: (params: ProcessListParams) => http.Get(`/process`, { params }),
  // 杀死进程 (保持兼容)
  kill: (pid: number) => http.Post(`/process/kill`, { pid }),
  // 发送信号
  signal: (params: ProcessSignalParams) => http.Post(`/process/signal`, params),
  // 获取进程详情
  detail: (pid: number) => http.Get(`/process/${pid}/detail`)
}