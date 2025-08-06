<script setup lang="ts">
import { NModal, NCard, NDescriptions, NDescriptionsItem, NTag, NTabs, NTabPane, NSpin, NCode } from 'naive-ui'
import { useGettext } from 'vue3-gettext'
import { ref, watch, computed } from 'vue'

import process from '@/api/panel/process'
import { formatBytes, formatDateTime, formatPercent } from '@/utils'

const { $gettext } = useGettext()

interface Props {
  show: boolean
  pid: number | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

const loading = ref(false)
const processDetail = ref<any>(null)

const show = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// 监听PID变化，获取进程详情
watch(() => props.pid, async (newPid) => {
  if (newPid && props.show) {
    await fetchProcessDetail(newPid)
  }
}, { immediate: true })

watch(() => props.show, async (newShow) => {
  if (newShow && props.pid) {
    await fetchProcessDetail(props.pid)
  }
})

const fetchProcessDetail = async (pid: number) => {
  loading.value = true
  try {
    const response = await process.detail(pid)
    processDetail.value = response
  } catch (error) {
    window.$message.error($gettext('Failed to fetch process details'))
  } finally {
    loading.value = false
  }
}

const formatStatus = (status: string) => {
  switch (status) {
    case 'R':
      return { text: $gettext('Running'), type: 'success' }
    case 'S':
      return { text: $gettext('Sleeping'), type: 'warning' }
    case 'T':
      return { text: $gettext('Stopped'), type: 'error' }
    case 'I':
      return { text: $gettext('Idle'), type: 'primary' }
    case 'Z':
      return { text: $gettext('Zombie'), type: 'error' }
    case 'W':
      return { text: $gettext('Waiting'), type: 'warning' }
    case 'L':
      return { text: $gettext('Locked'), type: 'info' }
    default:
      return { text: status, type: 'default' }
  }
}
</script>

<template>
  <n-modal
    v-model:show="show"
    preset="card"
    :title="$gettext('Process Details - PID %{ pid }', { pid: pid || 0 })"
    style="width: 80vw; max-width: 1200px"
    size="huge"
    :bordered="false"
    :segmented="false"
  >
    <n-spin :show="loading">
      <div v-if="processDetail" style="min-height: 400px">
        <n-tabs type="line" animated>
          <!-- 基础信息 -->
          <n-tab-pane name="basic" :tab="$gettext('Basic Info')">
            <n-descriptions :column="2" bordered>
              <n-descriptions-item :label="$gettext('PID')">
                {{ processDetail.pid }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Parent PID')">
                {{ processDetail.ppid }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Name')">
                {{ processDetail.name }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('User')">
                {{ processDetail.username }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Status')">
                <n-tag :type="formatStatus(processDetail.status).type">
                  {{ formatStatus(processDetail.status).text }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Running')">
                <n-tag :type="processDetail.is_running ? 'success' : 'error'">
                  {{ processDetail.is_running ? $gettext('Yes') : $gettext('No') }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Executable')">
                {{ processDetail.exe || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Working Directory')">
                {{ processDetail.cwd || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Terminal')">
                {{ processDetail.terminal || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Start Time')">
                {{ formatDateTime(processDetail.start_time) }}
              </n-descriptions-item>
            </n-descriptions>
          </n-tab-pane>

          <!-- 资源使用 -->
          <n-tab-pane name="resources" :tab="$gettext('Resources')">
            <n-descriptions :column="2" bordered>
              <n-descriptions-item :label="$gettext('CPU Usage')">
                {{ formatPercent(processDetail.cpu) }}%
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('CPU Time')">
                {{ processDetail.cpu_time || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Memory (RSS)')">
                {{ formatBytes(processDetail.rss) }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Memory (VMS)')">
                {{ formatBytes(processDetail.vms) }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Memory Percentage')">
                {{ processDetail.mem_percent ? formatPercent(processDetail.mem_percent) + '%' : '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Threads')">
                {{ processDetail.num_threads }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Priority (Nice)')">
                {{ processDetail.nice || '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('IO Priority')">
                {{ processDetail.ionice || '-' }}
              </n-descriptions-item>
            </n-descriptions>
          </n-tab-pane>

          <!-- 系统信息 -->
          <n-tab-pane name="system" :tab="$gettext('System Info')">
            <n-descriptions :column="2" bordered>
              <n-descriptions-item :label="$gettext('User IDs')">
                {{ processDetail.uids ? processDetail.uids.join(', ') : '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Group IDs')">
                {{ processDetail.gids ? processDetail.gids.join(', ') : '-' }}
              </n-descriptions-item>
              <n-descriptions-item :label="$gettext('Child Processes')">
                {{ processDetail.children ? processDetail.children.join(', ') : $gettext('None') }}
              </n-descriptions-item>
            </n-descriptions>
          </n-tab-pane>

          <!-- 命令行 -->
          <n-tab-pane name="cmdline" :tab="$gettext('Command Line')">
            <n-code
              :code="processDetail.cmdline ? processDetail.cmdline.join(' ') : $gettext('No command line available')"
              language="bash"
              show-line-numbers
              word-wrap
            />
          </n-tab-pane>

          <!-- 环境变量 -->
          <n-tab-pane name="env" :tab="$gettext('Environment')" v-if="processDetail.environ">
            <n-code
              :code="Object.entries(processDetail.environ || {}).map(([key, value]) => `${key}=${value}`).join('\n')"
              language="bash"
              show-line-numbers
              word-wrap
            />
          </n-tab-pane>
        </n-tabs>
      </div>
    </n-spin>
  </n-modal>
</template>