<script setup lang="ts">
import { reactive, ref, computed, onMounted, nextTick, h } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NDropdown,
  NEmpty,
  NFlex,
  NIcon,
  NInput,
  NModal,
  NPopconfirm,
  NSelect,
  NSpace,
  NTag,
  NTree,
  NTabs,
  NTabPane,
  NDescriptions,
  NDescriptionsItem,
  NGrid,
  NGridItem,
  NProgress,
  useMessage,
  useDialog,
  NTooltip
} from 'naive-ui'
import { useGettext } from 'vue3-gettext'
import { 
  Search, 
  Refresh, 
  TreeOutline, 
  List,
  InformationCircle,
  StopCircle,
  PlayCircle,
  Warning,
  ChevronDown,
  Terminal
} from '@vicons/ionicons5'

import process from '@/api/panel/process'
import { formatBytes, formatDateTime, formatPercent, renderIcon } from '@/utils'

const { $gettext } = useGettext()
const message = useMessage()
const dialog = useDialog()

// 响应式数据
const loading = ref(false)
const treeLoading = ref(false)
const viewMode = ref<'list' | 'tree'>('list')
const data = ref<any[]>([])
const treeData = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const pageCount = computed(() => Math.ceil(total.value / pageSize.value))

// 筛选和搜索
const filters = reactive({
  search: '',
  status: '',
  username: '',
  sortBy: 'pid',
  sortDir: 'asc'
})

// 进程详情模态框
const detailModal = ref(false)
const detailData = ref<any>({})
const detailLoading = ref(false)

// 信号选项
const signalOptions = [
  { label: 'SIGTERM (温和终止)', value: 'SIGTERM' },
  { label: 'SIGKILL (强制杀死)', value: 'SIGKILL' },
  { label: 'SIGINT (中断)', value: 'SIGINT' },
  { label: 'SIGHUP (挂起)', value: 'SIGHUP' },
  { label: 'SIGUSR1 (用户信号1)', value: 'SIGUSR1' },
  { label: 'SIGUSR2 (用户信号2)', value: 'SIGUSR2' },
  { label: 'SIGQUIT (退出)', value: 'SIGQUIT' },
  { label: 'SIGSTOP (暂停)', value: 'SIGSTOP' },
  { label: 'SIGCONT (继续)', value: 'SIGCONT' }
]

// 状态选项
const statusOptions = [
  { label: $gettext('All'), value: '' },
  { label: $gettext('Running'), value: 'R' },
  { label: $gettext('Sleeping'), value: 'S' },
  { label: $gettext('Idle'), value: 'I' },
  { label: $gettext('Zombie'), value: 'Z' },
  { label: $gettext('Waiting'), value: 'D' },
  { label: $gettext('Locked'), value: 'L' }
]

// 排序选项
const sortOptions = [
  { label: 'PID', value: 'pid' },
  { label: $gettext('Name'), value: 'name' },
  { label: 'CPU', value: 'cpu' },
  { label: $gettext('Memory'), value: 'memory' },
  { label: $gettext('Start Time'), value: 'start_time' }
]

// 表格列定义
const columns: any = [
  {
    title: 'PID',
    key: 'pid',
    width: 80,
    ellipsis: { tooltip: true },
    render: (row: any) => {
      return h(NButton, {
        text: true,
        type: 'primary',
        onClick: () => showDetail(row.pid)
      }, { default: () => row.pid })
    }
  },
  {
    title: $gettext('Name'),
    key: 'name',
    minWidth: 200,
    resizable: true,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Parent PID'),
    key: 'ppid',
    width: 100,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('User'),
    key: 'username',
    minWidth: 120,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Status'),
    key: 'status',
    width: 100,
    render: (row: any) => {
      const statusMap: any = {
        'R': { type: 'success', text: $gettext('Running') },
        'S': { type: 'info', text: $gettext('Sleeping') },
        'I': { type: 'info', text: $gettext('Idle') },
        'Z': { type: 'warning', text: $gettext('Zombie') },
        'D': { type: 'warning', text: $gettext('Waiting') },
        'L': { type: 'error', text: $gettext('Locked') }
      }
      const status = statusMap[row.status] || { type: 'default', text: row.status }
      return h(NTag, { type: status.type }, { default: () => status.text })
    }
  },
  {
    title: 'CPU %',
    key: 'cpu',
    width: 100,
    render: (row: any) => formatPercent(row.cpu)
  },
  {
    title: $gettext('Memory'),
    key: 'rss',
    width: 120,
    render: (row: any) => formatBytes(row.rss)
  },
  {
    title: $gettext('Memory') + ' %',
    key: 'memory_percent', 
    width: 100,
    render: (row: any) => formatPercent(row.memory_percent)
  },
  {
    title: $gettext('Start Time'),
    key: 'start_time',
    width: 160,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Command'),
    key: 'cmd_line',
    minWidth: 300,
    ellipsis: { tooltip: true }
  },
  {
    title: $gettext('Actions'),
    key: 'actions',
    width: 120,
    render: (row: any) => {
      return h(NSpace, {}, {
        default: () => [
          h(NDropdown, {
            trigger: 'click',
            options: signalOptions.map(opt => ({
              label: opt.label,
              key: opt.value,
              props: {
                onClick: () => sendSignal(row.pid, opt.value)
              }
            }))
          }, {
            default: () => h(NButton, { size: 'small', secondary: true }, {
              default: () => [$gettext('Signal'), h(NIcon, { component: ChevronDown })]
            })
          }),
          h(NPopconfirm, {
            onPositiveClick: () => killProcess(row.pid)
          }, {
            trigger: () => h(NButton, { 
              size: 'small', 
              type: 'error',
              secondary: true
            }, {
              default: () => [h(NIcon, { component: StopCircle }), $gettext('Kill')]
            }),
            default: () => $gettext('Are you sure you want to terminate process %{ pid }?', { pid: row.pid })
          })
        ]
      })
    }
  }
]

// 树形节点渲染
const renderTreeNode = ({ option }: any) => {
  return h('div', { 
    style: { 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'space-between',
      width: '100%'
    } 
  }, [
    h('span', { 
      style: { flex: 1 }
    }, `${option.pid} - ${option.name} (${option.username})`),
    h(NSpace, { size: 'small' }, [
      h(NTag, { size: 'small', type: 'info' }, option.status),
      h('span', { style: { fontSize: '12px', color: '#999' } }, formatPercent(option.cpu)),
      h('span', { style: { fontSize: '12px', color: '#999' } }, formatBytes(option.rss))
    ])
  ])
}

// 获取进程列表
const getProcessList = async () => {
  loading.value = true
  try {
    const { data: result } = await process.list({
      page: page.value,
      limit: pageSize.value,
      sort_by: filters.sortBy,
      sort_dir: filters.sortDir,
      status: filters.status,
      search: filters.search,
      username: filters.username
    })
    data.value = result.items || []
    total.value = result.total || 0
  } catch (error: any) {
    message.error(error.message || '获取进程列表失败')
  }
  loading.value = false
}

// 获取进程树
const getProcessTree = async () => {
  treeLoading.value = true
  try {
    const { data: result } = await process.tree()
    treeData.value = transformTreeData(result || [])
  } catch (error: any) {
    message.error(error.message || '获取进程树失败') 
  }
  treeLoading.value = false
}

// 转换树形数据
const transformTreeData = (nodes: any[]): any[] => {
  return nodes.map(node => ({
    key: node.pid,
    label: `${node.pid} - ${node.name}`,
    pid: node.pid,
    name: node.name,
    username: node.username,
    status: node.status,
    cpu: node.cpu,
    rss: node.rss,
    children: node.children ? transformTreeData(node.children) : undefined
  }))
}

// 显示进程详情
const showDetail = async (pid: number) => {
  detailModal.value = true
  detailLoading.value = true
  try {
    const { data: result } = await process.detail(pid)
    detailData.value = result
  } catch (error: any) {
    message.error(error.message || '获取进程详情失败')
  }
  detailLoading.value = false
}

// 发送信号
const sendSignal = async (pid: number, signal: string) => {
  try {
    await process.signal(pid, signal)
    message.success($gettext('Signal sent successfully'))
    await getProcessList()
  } catch (error: any) {
    message.error(error.message || '发送信号失败')
  }
}

// 杀死进程
const killProcess = async (pid: number) => {
  try {
    await process.kill(pid)
    message.success($gettext('Process %{ pid } has been terminated', { pid }))
    await getProcessList()
  } catch (error: any) {
    message.error(error.message || '终止进程失败')
  }
}

// 刷新数据
const refresh = () => {
  if (viewMode.value === 'list') {
    getProcessList()
  } else {
    getProcessTree()
  }
}

// 切换视图模式
const switchViewMode = (mode: 'list' | 'tree') => {
  viewMode.value = mode
  nextTick(() => {
    if (mode === 'list') {
      getProcessList()
    } else {
      getProcessTree()
    }
  })
}

// 重置筛选
const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  filters.username = ''
  filters.sortBy = 'pid'
  filters.sortDir = 'asc'
  page.value = 1
  getProcessList()
}

// 监听筛选变化
const onFiltersChange = () => {
  page.value = 1
  getProcessList()
}

// 组件挂载
onMounted(() => {
  getProcessList()
})
</script>

<template>
  <div class="process-manager">
    <!-- 工具栏 -->
    <NCard class="mb-4">
      <NFlex justify="space-between" align="center">
        <NFlex align="center">
          <!-- 视图切换 -->
          <NSpace>
            <NButton
              :type="viewMode === 'list' ? 'primary' : 'default'"
              @click="switchViewMode('list')"
            >
              <template #icon>
                <NIcon :component="List" />
              </template>
              {{ $gettext('List View') }}
            </NButton>
            <NButton
              :type="viewMode === 'tree' ? 'primary' : 'default'"
              @click="switchViewMode('tree')"
            >
              <template #icon>
                <NIcon :component="TreeOutline" />
              </template>
              {{ $gettext('Tree View') }}
            </NButton>
          </NSpace>

          <!-- 搜索 -->
          <NInput
            v-model:value="filters.search"
            :placeholder="$gettext('Search by PID, name or command...')"
            style="width: 300px"
            clearable
            @update:value="onFiltersChange"
          >
            <template #prefix>
              <NIcon :component="Search" />
            </template>
          </NInput>
        </NFlex>

        <!-- 筛选和刷新 -->
        <NSpace>
          <NSelect
            v-model:value="filters.status"
            :options="statusOptions"
            :placeholder="$gettext('Filter by status')"
            style="width: 150px"
            clearable
            @update:value="onFiltersChange"
          />
          
          <NSelect
            v-model:value="filters.sortBy"
            :options="sortOptions"
            :placeholder="$gettext('Sort by')"
            style="width: 120px"
            @update:value="onFiltersChange"
          />
          
          <NSelect
            v-model:value="filters.sortDir"
            :options="[
              { label: $gettext('Ascending'), value: 'asc' },
              { label: $gettext('Descending'), value: 'desc' }
            ]"
            style="width: 100px"
            @update:value="onFiltersChange"
          />

          <NButton secondary @click="resetFilters">
            {{ $gettext('Reset') }}
          </NButton>

          <NButton type="primary" @click="refresh" :loading="loading || treeLoading">
            <template #icon>
              <NIcon :component="Refresh" />
            </template>
            {{ $gettext('Refresh') }}
          </NButton>
        </NSpace>
      </NFlex>
    </NCard>

    <!-- 进程列表视图 -->
    <NCard v-if="viewMode === 'list'">
      <NDataTable
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
          pageSizes: [20, 50, 100, 200],
          onUpdatePage: getProcessList,
          onUpdatePageSize: () => {
            page = 1
            getProcessList()
          }
        }"
      />
    </NCard>

    <!-- 进程树视图 -->
    <NCard v-if="viewMode === 'tree'">
      <div v-if="treeLoading" class="text-center py-8">
        <NIcon size="32" :component="TreeOutline" class="animate-spin" />
        <div class="mt-2">{{ $gettext('Loading process tree...') }}</div>
      </div>
      <NTree
        v-else-if="treeData.length > 0"
        :data="treeData"
        block-line
        expand-on-click
        :render-label="renderTreeNode"
        :on-load="() => Promise.resolve()"
      />
      <NEmpty v-else :description="$gettext('No processes found')" />
    </NCard>

    <!-- 进程详情模态框 -->
    <NModal v-model:show="detailModal" preset="card" style="width: 80%; max-width: 1200px" :title="$gettext('Process Details')">
      <div v-if="detailLoading" class="text-center py-8">
        {{ $gettext('Loading...') }}
      </div>
      <div v-else-if="detailData.pid">
        <NTabs>
          <NTabPane name="basic" :tab="$gettext('Basic Info')">
            <NGrid :cols="2" :x-gap="24" :y-gap="16">
              <NGridItem>
                <NDescriptions bordered>
                  <NDescriptionsItem :label="'PID'">{{ detailData.pid }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Name')">{{ detailData.name }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Parent PID')">{{ detailData.ppid }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('User')">{{ detailData.username }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Status')">
                    <NTag :type="detailData.status === 'R' ? 'success' : 'info'">{{ detailData.status }}</NTag>
                  </NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Start Time')">{{ detailData.start_time }}</NDescriptionsItem>
                </NDescriptions>
              </NGridItem>
              <NGridItem>
                <NDescriptions bordered>
                  <NDescriptionsItem :label="$gettext('Executable')">{{ detailData.exe || 'N/A' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Working Directory')">{{ detailData.cwd || 'N/A' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Terminal')">{{ detailData.terminal || 'N/A' }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Priority')">{{ detailData.nice }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('Threads')">{{ detailData.num_threads }}</NDescriptionsItem>
                  <NDescriptionsItem :label="$gettext('File Descriptors')">{{ detailData.fds || 'N/A' }}</NDescriptionsItem>
                </NDescriptions>
              </NGridItem>
            </NGrid>
          </NTabPane>

          <NTabPane name="resources" :tab="$gettext('Resources')">
            <NGrid :cols="2" :x-gap="24" :y-gap="16">
              <NGridItem>
                <NCard :title="$gettext('CPU Usage')">
                  <NProgress type="circle" :percentage="detailData.cpu" />
                  <div class="mt-4">
                    <div>{{ $gettext('CPU Percent') }}: {{ formatPercent(detailData.cpu) }}</div>
                  </div>
                </NCard>
              </NGridItem>
              <NGridItem>
                <NCard :title="$gettext('Memory Usage')">
                  <NProgress type="circle" :percentage="detailData.memory_percent" />
                  <div class="mt-4">
                    <div>{{ $gettext('RSS') }}: {{ formatBytes(detailData.rss) }}</div>
                    <div>{{ $gettext('VMS') }}: {{ formatBytes(detailData.vms) }}</div>
                    <div>{{ $gettext('Memory Percent') }}: {{ formatPercent(detailData.memory_percent) }}</div>
                  </div>
                </NCard>
              </NGridItem>
            </NGrid>
          </NTabPane>

          <NTabPane name="command" :tab="$gettext('Command Line')">
            <NCard>
              <div class="font-mono text-sm whitespace-pre-wrap bg-gray-50 p-4 rounded">
                {{ detailData.cmd_line || 'N/A' }}
              </div>
              <div v-if="detailData.command_line && detailData.command_line.length" class="mt-4">
                <div class="text-sm font-medium mb-2">{{ $gettext('Arguments') }}:</div>
                <div class="space-y-1">
                  <div v-for="(arg, index) in detailData.command_line" :key="index" class="font-mono text-sm bg-gray-50 p-2 rounded">
                    [{{ index }}] {{ arg }}
                  </div>
                </div>
              </div>
            </NCard>
          </NTabPane>

          <NTabPane name="children" :tab="$gettext('Child Processes')" v-if="detailData.children_detail && detailData.children_detail.length">
            <NDataTable
              :columns="[
                { title: 'PID', key: 'pid', width: 80 },
                { title: $gettext('Name'), key: 'name', minWidth: 200 },
                { title: $gettext('Status'), key: 'status', width: 100 },
                { title: 'CPU %', key: 'cpu', width: 100, render: (row: any) => formatPercent(row.cpu) },
                { title: $gettext('Memory'), key: 'rss', width: 120, render: (row: any) => formatBytes(row.rss) }
              ]"
              :data="detailData.children_detail"
              :max-height="400"
            />
          </NTabPane>
        </NTabs>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.process-manager {
  padding: 0 24px;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
