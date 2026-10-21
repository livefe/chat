import { ref } from 'vue'

import { getImContacts, type ImContactItem } from '@/api/im'
import { getSupportOverviewSummary, type SupportAdminSession } from '@/api/support'
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

export function useSupportOverviewFilters() {
  const keywordDraft = ref('')
  const keyword = ref('')
  const assigneeId = ref<number | undefined>()
  const agentOptions = ref<ImContactItem[]>([])

  function applyKeyword() {
    keyword.value = keywordDraft.value.trim()
  }

  function listFilterParams() {
    return {
      assigneeEmployeeId: assigneeId.value || undefined,
      keyword: keyword.value.trim() || undefined,
    }
  }

  function filterQueue(list: SupportAdminSession[]) {
    const kw = keyword.value.trim().toLowerCase()
    if (!kw) return list
    return list.filter((s) => {
      const hay = `${s.userId} ${s.id} ${s.userMobile || ''} ${s.userNickname || ''}`.toLowerCase()
      return hay.includes(kw)
    })
  }

  function mergeAgentOptions(list: SupportAdminSession[]) {
    const map = new Map(agentOptions.value.map((x) => [x.id, x]))
    for (const s of list) {
      if (!s.assigneeEmployeeId || map.has(s.assigneeEmployeeId)) continue
      map.set(s.assigneeEmployeeId, {
        id: s.assigneeEmployeeId,
        name: s.assigneeName || `坐席#${s.assigneeEmployeeId}`,
        avatarUrl: '',
        departmentId: 0,
        departmentName: '',
        online: false,
      })
    }
    agentOptions.value = [...map.values()].sort((a, b) => a.id - b.id)
  }

  async function loadAgents() {
    try {
      const { data } = await getImContacts({ page: 1, pageSize: 200 })
      const map = new Map(agentOptions.value.map((x) => [x.id, x]))
      for (const c of data.list || []) map.set(c.id, c)
      agentOptions.value = [...map.values()].sort((a, b) => a.id - b.id)
    } catch {
      return
    }
  }

  return {
    keywordDraft,
    assigneeId,
    agentOptions,
    applyKeyword,
    listFilterParams,
    filterQueue,
    mergeAgentOptions,
    loadAgents,
  }
}
