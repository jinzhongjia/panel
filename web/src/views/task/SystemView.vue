<script setup lang="ts">
import { NButton, NDataTable, NPopconfirm, NTag, NInput, NSelect, NSwitch, NDropdown, NSpace, NFlex } from 'naive-ui'
import { useGettext } from 'vue3-gettext'
import { ref, computed, watch, h } from 'vue'

import process from '@/api/panel/process'
import { formatBytes, formatDateTime, formatPercent, renderIcon } from '@/utils'
import ProcessDetailModal from './ProcessDetailModal.vue'

const { $gettext } = useGettext()

// 搜索和过滤状态
const searchText = ref('')
const statusFilter = ref('')
const sortBy = ref('cpu')
const sortDesc = ref(true)
const showTree = ref(false)

// 状态选项
const statusOptions = [
  { label: $gettext('All'), value: '' },
  { label: $gettext('Running'), value: 'running' },
  { label: $gettext('Sleeping'), value: 'sleeping' },
  { label: $gettext('Stopped'), value: 'stopped' },
  { label: $gettext('Zombie'), value: 'zombie' },
  { label: $gettext('Idle'), value: 'idle' }
]

// 排序选项
const sortOptions = [
  { label: 'PID', value: 'pid' },
  { label: $gettext('Name'), value: 'name' },
  { label: 'CPU', value: 'cpu' },
  { label: $gettext('Memory'), value: 'memory' },
  { label: $gettext('Start Time'), value: 'start_time' }
]

// 信号选项
const signalOptions = [
  { label: 'SIGTERM - ' + $gettext('Graceful termination'), key: 'SIGTERM' },
  { label: 'SIGKILL - ' + $gettext('Force kill'), key: 'SIGKILL' },
  { label: 'SIGINT - ' + $gettext('Interrupt'), key: 'SIGINT' },
  { label: 'SIGQUIT - ' + $gettext('Quit'), key: 'SIGQUIT' },
  { label: 'SIGSTOP - ' + $gettext('Pause process'), key: 'SIGSTOP' },
  { label: 'SIGCONT - ' + $gettext('Continue process'), key: 'SIGCONT' },
  { label: 'SIGHUP - ' + $gettext('Hangup'), key: 'SIGHUP' },
  { label: 'SIGUSR1 - ' + $gettext('User signal 1'), key: 'SIGUSR1' },
  { label: 'SIGUSR2 - ' + $gettext('User signal 2'), key: 'SIGUSR2' }
]

// 发送信号处理
const handleSendSignal = (pid: number, signal: string) => {
  useRequest(process.signal({ pid, signal })).onSuccess(() => {
    refresh()
    window.$message.success($gettext('Signal %{ signal } sent to process %{ pid }', { signal, pid }))
  })
}

// 进程详情弹窗状态
const showDetailModal = ref(false)
const selectedPid = ref<number | null>(null)

// 查看进程详情
const handleViewDetail = (pid: number) => {
  selectedPid.value = pid
  showDetailModal.value = true
}
const columns: any = [
  {
    title: 'PID',
    key: 'pid',
    width: 120,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Name'),
    key: 'name',
    minWidth: 250,
    resizable: true,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Parent PID'),
    key: 'ppid',
    width: 120,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Threads'),
    key: 'num_threads',
    width: 100,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('User'),
    key: 'username',
    minWidth: 100,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Status'),
    key: 'status',
    minWidth: 150,
    ellipsis: { tooltip: true },
    render(row: any) {
      switch (row.status) {
        case 'R':
          return h(NTag, { type: 'success' }, { default: () => $gettext('Running') })
        case 'S':
          return h(NTag, { type: 'warning' }, { default: () => $gettext('Sleeping') })
        case 'T':
          return h(NTag, { type: 'error' }, { default: () => $gettext('Stopped') })
        case 'I':
          return h(NTag, { type: 'primary' }, { default: () => $gettext('Idle') })
        case 'Z':
          return h(NTag, { type: 'error' }, { default: () => $gettext('Zombie') })
        case 'W':
          return h(NTag, { type: 'warning' }, { default: () => $gettext('Waiting') })
        case 'L':
          return h(NTag, { type: 'info' }, { default: () => $gettext('Locked') })
        default:
          return h(NTag, { type: 'default' }, { default: () => row.status })
      }
    }
  },
  {
    title: 'CPU',
    key: 'cpu',
    minWidth: 100,
    ellipsis: { tooltip: true },
    render(row: any): string {
      return formatPercent(row.cpu) + '%'
    }
  },
  {
    title: $gettext('Memory'),
    key: 'rss',
    minWidth: 100,
    ellipsis: { tooltip: true },
    render(row: any): string {
      return formatBytes(row.rss)
    }
  },
  {
    title: $gettext('Memory %'),
    key: 'mem_percent',
    width: 100,
    ellipsis: { tooltip: true },
    render(row: any): string {
      return row.mem_percent ? formatPercent(row.mem_percent) + '%' : '-'
    }
  },
  {
    title: $gettext('Start Time'),
    key: 'start_time',
    width: 160,
    ellipsis: { tooltip: true },
    render(row: any): string {
      return formatDateTime(row.start_time)
    }
  },
  {
    title: $gettext('Actions'),
    key: 'actions',
    width: 200,
    hideInExcel: true,
    render(row: any) {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          // 查看详情按钮
          h(NButton, {
            size: 'small',
            type: 'info',
            onClick: () => handleViewDetail(row.pid)
          }, {
            default: () => $gettext('Detail'),
            icon: renderIcon('material-symbols:info-outline', { size: 14 })
          }),
          // 信号下拉菜单
          h(NDropdown, {
            trigger: 'click',
            options: signalOptions.map(option => ({
              label: option.label,
              key: option.key,
              props: {
                onClick: () => {
                  if (option.key === 'SIGKILL' || option.key === 'SIGTERM') {
                    // 对于危险操作显示确认对话框
                    window.$dialog.warning({
                      title: $gettext('Confirm Signal'),
                      content: $gettext('Are you sure you want to send %{ signal } to process %{ pid }?', {
                        signal: option.key,
                        pid: row.pid
                      }),
                      positiveText: $gettext('Confirm'),
                      negativeText: $gettext('Cancel'),
                      onPositiveClick: () => handleSendSignal(row.pid, option.key)
                    })
                  } else {
                    handleSendSignal(row.pid, option.key)
                  }
                }
              }
            }))
          }, {
            default: () => h(NButton, {
              size: 'small',
              type: 'warning'
            }, {
              default: () => $gettext('Signal'),
              icon: renderIcon('material-symbols:send', { size: 14 })
            })
          }),
          // 快速终止按钮 (保持兼容)
          h(NPopconfirm, {
            onPositiveClick: () => {
              useRequest(process.kill(row.pid)).onSuccess(() => {
                refresh()
                window.$message.success(
                  $gettext('Process %{ pid } has been terminated', { pid: row.pid })
                )
              })
            }
          }, {
            default: () => {
              return $gettext('Are you sure you want to terminate process %{ pid }?', {
                pid: row.pid
              })
            },
            trigger: () => {
              return h(NButton, {
                size: 'small',
                type: 'error'
              }, {
                default: () => $gettext('Kill'),
                icon: renderIcon('material-symbols:stop-circle-outline-rounded', { size: 14 })
              })
            }
          })
        ]
      })
    }
  }
]
// 计算查询参数
const queryParams = computed(() => ({
  page: page.value,
  limit: pageSize.value,
  sort_by: sortBy.value,
  sort_desc: sortDesc.value,
  status: statusFilter.value || undefined,
  search: searchText.value || undefined,
  show_tree: showTree.value
}))

const { loading, data, page, total, pageSize, pageCount, refresh } = usePagination(
  () => process.list(queryParams.value),
  {
    initialData: { total: 0, list: [] },
    initialPageSize: 20,
    total: (res: any) => res.total,
    data: (res: any) => res.items
  }
)

// 监听搜索和过滤条件变化
watch([searchText, statusFilter, sortBy, sortDesc, showTree], () => {
  page.value = 1
  refresh()
}, { debounce: 300 })</script>

<template>
  <n-flex vertical>
    <!-- 搜索和过滤工具栏 -->
    <n-flex justify="space-between" align="center">
      <n-flex align="center" :size="16">
        <n-input
          v-model:value="searchText"
          :placeholder="$gettext('Search by name, command, user, or PID')"
          clearable
          style="width: 300px"
        >
          <template #prefix>
            <the-icon icon="material-symbols:search" />
          </template>
        </n-input>
        
        <n-select
          v-model:value="statusFilter"
          :options="statusOptions"
          :placeholder="$gettext('Filter by status')"
          clearable
          style="width: 150px"
        />
        
        <n-select
          v-model:value="sortBy"
          :options="sortOptions"
          :placeholder="$gettext('Sort by')"
          style="width: 120px"
        />
        
        <n-flex align="center" :size="8">
          <span>{{ $gettext('Desc') }}</span>
          <n-switch v-model:value="sortDesc" />
        </n-flex>
      </n-flex>
      
      <n-flex align="center" :size="16">
        <n-flex align="center" :size="8">
          <span>{{ $gettext('Tree View') }}</span>
          <n-switch v-model:value="showTree" />
        </n-flex>
        
        <n-button @click="refresh" :loading="loading">
          <template #icon>
            <the-icon icon="material-symbols:refresh" />
          </template>
          {{ $gettext('Refresh') }}
        </n-button>
      </n-flex>
    </n-flex>

    <!-- 进程表格 -->
    <n-data-table
      striped
      remote
      :scroll-x="1600"
      :loading="loading"
      :columns="columns"
      :data="data"
      :row-key="(row: any) => row.pid"
      v-model:page="page"
      v-model:pageSize="pageSize"
      :pagination="{
        page: page,
        pageCount: pageCount,
        pageSize: pageSize,
        itemCount: total,
        showQuickJumper: true,
        showSizePicker: true,
        pageSizes: [20, 50, 100, 200]
      }"
    />

    <!-- 进程详情弹窗 -->
    <ProcessDetailModal
      v-model:show="showDetailModal"
      :pid="selectedPid"
    />
  </n-flex>
</template>