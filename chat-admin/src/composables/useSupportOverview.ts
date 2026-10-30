import { ref } from 'vue'

import {
  getSupportAgents,
  getSupportOverviewSummary,
  type SupportAgent,
  type SupportOverviewParams,
} from '@/api/support'
import { showApiError } from '@/utils/apiError'

export function useSupportOverviewSummary() {
  const summary = ref({ queueCount: 0, activeCount: 0, longWaitQueueCount: 0 })

  async function loadSummary(silent = false) {
    try {
      const { data } = await getSupportOverviewSummary()
      summary.value = {
        queueCount: data.queueCount || 0,
        activeCount: data.activeCount || 0,
        longWaitQueueCount: data.longWaitQueueCount || 0,
      }
    } catch (e) {
      if (!silent) showApiError(e, '加载汇总失败')
    }
  }

  return { summary, loadSummary }
}

export function useSupportAgents() {
  const agents = ref<SupportAgent[]>([])
  const agentsLoading = ref(false)

  async function loadAgents() {
    agentsLoading.value = true
    try {
      const { data } = await getSupportAgents()
      agents.value = [...(data.list || [])].sort(
        (a, b) => Number(b.online) - Number(a.online) || a.id - b.id,
      )
    } catch (e) {
      showApiError(e, '加载坐席失败')
    } finally {
      agentsLoading.value = false
    }
  }

  return { agents, agentsLoading, loadAgents }
}

export interface SupportOverviewFilters {
  keyword: string
  assigneeId?: number
  dateRange: [string, string] | null
}

export function overviewListParams(
  f: SupportOverviewFilters,
): Omit<SupportOverviewParams, 'page' | 'pageSize'> {
  return {
    keyword: f.keyword.trim() || undefined,
    assigneeEmployeeId: f.assigneeId || undefined,
    startDate: f.dateRange?.[0] || undefined,
    endDate: f.dateRange?.[1] || undefined,
  }
}
