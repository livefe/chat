<template>
  <div class="admin-page im-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="canOpenConfig" class="card-header-actions">
            <el-button @click="configVisible = true">
              <el-icon><Setting /></el-icon>
              客服配置
            </el-button>
          </div>
        </div>
      </template>

      <ImConnectionBanner :state="wsState" @reconnect="reconnect" />

      <ImShell :active-tab="sideTab" :tabs="tabs" @update:active-tab="onTabChange">
        <template #search>
          <SupportOverviewSearch
            v-model:keyword="keywordDraft"
            v-model:assignee-id="filters.assigneeId"
            v-model:date-range="filters.dateRange"
            :show-advanced="sideTab !== 'queue'"
            :date-label="sideTab === 'closed' ? '结案' : '接入'"
            :agent-options="agents"
            @search="applyFilters"
          />
        </template>

        <template #list>
          <div v-loading="listLoading">
            <SupportOverviewSessionLists
              :tab="sideTab"
              :queue="queue"
              :active-sessions="activeSessions"
              :closed-sessions="closedSessions"
              :active-id="active?.id"
              :now="now"
              @select="selectSession"
            />
          </div>
        </template>

        <template #main>
          <template v-if="active">
            <ImChatHeader>
              <template #title>
                <UserCell
                  :user-id="active.userId"
                  :nickname="active.userNickname"
                  :mobile="active.userMobile"
                />
              </template>
              <template #tags>
                <el-tag v-if="isQueuedActive" size="small" type="warning" effect="plain">
                  排队中 · 已等待 {{ imFormatWait(supportWaitSec(active, now)) }}
                </el-tag>
                <el-tag v-else-if="isClosedActive" size="small" type="info" effect="plain">
                  已结束
                </el-tag>
              </template>
              <template v-if="headerSub" #sub>{{ headerSub }}</template>
              <template #actions>
                <el-button
                  v-if="showAccept"
                  type="primary"
                  :loading="accepting"
                  @click="onAccept(active)"
                >
                  接入
                </el-button>
                <el-button v-if="showTransferToSelf" type="primary" @click="onTransferToSelf">
                  转接给我
                </el-button>
                <el-button v-if="showPeerTransfer" @click="transferVisible = true">
                  转接给同事
                </el-button>
                <el-button v-if="showGoDesk" @click="goSupportDesk">去工作台</el-button>
              </template>
            </ImChatHeader>
            <ImMessageThread
              ref="threadRef"
              :messages="messages"
              :self-id="selfId"
              mine-mode="support"
              :loading="messagesLoading"
              :has-more="messagesHasMore"
              :loading-more="messagesLoadingMore"
              @load-more="loadMoreMessages"
            />
            <div class="im-closed-bar">{{ footerText }}</div>
          </template>
          <div v-else class="im-empty">
            <div class="im-empty__title">{{ route.meta.title }}</div>
            <div class="im-empty__desc">盯排队与全量会话；他人会话只读，介入请转接</div>
          </div>
        </template>
      </ImShell>
    </el-card>

    <SupportTransferDialog
      v-model="transferVisible"
      title="转接给同事"
      :conversation-id="active?.id"
      :exclude-ids="[selfId, Number(active?.assigneeEmployeeId || 0)]"
      @transferred="onTransferred"
    />

    <el-drawer v-model="configVisible" title="客服配置" size="560px" destroy-on-close>
      <SupportConfig />
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter, type LocationQueryValue } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Setting } from '@element-plus/icons-vue'
import { useEmployeeStore } from '@/stores/employee'
import {
  acceptSupport,
  getSupportActiveAll,
  getSupportClosedAll,
  getSupportQueue,
  getSupportSession,
  transferSupport,
  type SupportAdminSession,
} from '@/api/support'
import { PERM } from '@/constants/permission'
import { SupportSessionStatus } from '@/constants/support'
import { usePermission } from '@/composables/usePermission'
import { useSupportMessaging } from '@/composables/useSupportMessaging'
import {
  overviewListParams,
  useSupportAgents,
  useSupportOverviewSummary,
  type SupportOverviewFilters,
} from '@/composables/useSupportOverview'
import {
  reselectActiveSession,
  supportSessionTitle,
  supportWaitSec,
  useNowSec,
  useSupportSessionList,
} from '@/composables/useSupportSessions'
import { showApiError } from '@/utils/apiError'
import { usePageVisible } from '@/utils/attention'
import { isCancelAction } from '@/utils/confirm'
import { findMenuPath } from '@/utils/navigation'
import UserCell from '@/components/common/UserCell.vue'
import ImConnectionBanner from '@/components/im/ImConnectionBanner.vue'
import ImShell from '@/components/im/ImShell.vue'
import ImChatHeader from '@/components/im/ImChatHeader.vue'
import ImMessageThread from '@/components/im/ImMessageThread.vue'
import { imFormatWait } from '@/components/im/media'
import SupportConfig from './SupportConfig.vue'
import SupportOverviewSearch from './SupportOverviewSearch.vue'
import SupportOverviewSessionLists from './SupportOverviewSessionLists.vue'
import SupportTransferDialog from './SupportTransferDialog.vue'

const LIVE_POLL_MS = 5000

type SideTab = 'queue' | 'active' | 'closed'
const SIDE_TABS: SideTab[] = ['queue', 'active', 'closed']

const employeeStore = useEmployeeStore()
const { can } = usePermission()
const route = useRoute()
const router = useRouter()
const now = useNowSec()
const pageVisible = usePageVisible()
const selfId = computed(() => employeeStore.employeeInfo?.id || 0)

const canAccept = computed(() => can(PERM.ImSupportAccept))
const canTransfer = computed(() => can(PERM.ImSupportTransfer))
const canOpenConfig = computed(() => can(PERM.ImSupportOverviewConfig))

const configVisible = ref(false)
const sideTab = ref<SideTab>('queue')
const active = ref<SupportAdminSession | null>(null)
const transferVisible = ref(false)
const accepting = ref(false)
const keywordDraft = ref('')
const filters = reactive<SupportOverviewFilters>({
  keyword: '',
  assigneeId: undefined,
  dateRange: null,
})

const { summary, loadSummary } = useSupportOverviewSummary()
const { agents, loadAgents } = useSupportAgents()

const queue = useSupportSessionList({
  fetch: (p) => getSupportQueue({ ...p, keyword: filters.keyword || undefined }),
  errorText: '加载排队失败',
  onReplaced: (list) =>
    reselectActiveSession(active, list, (s) => s.status === SupportSessionStatus.Queued),
})
const activeSessions = useSupportSessionList({
  fetch: (p) => getSupportActiveAll({ ...p, ...overviewListParams(filters) }),
  errorText: '加载进行中失败',
  onReplaced: (list) =>
    reselectActiveSession(active, list, (s) => s.status === SupportSessionStatus.Active),
})
const closedSessions = useSupportSessionList({
  fetch: (p) => getSupportClosedAll({ ...p, ...overviewListParams(filters) }),
  errorText: '加载已结案失败',
  onReplaced: (list) =>
    reselectActiveSession(active, list, (s) => s.status === SupportSessionStatus.Closed),
})

const {
  messages,
  messagesLoading,
  messagesLoadingMore,
  messagesHasMore,
  threadRef,
  wsState,
  reconnect,
  start,
  loadMessages,
  loadMoreMessages,
  softRefreshMessages,
  clearMessages,
} = useSupportMessaging(active, {
  unreadList: activeSessions,
  previewLists: [queue],
  refreshLists: softRefreshAll,
  canMarkRead: (s) =>
    s.status === SupportSessionStatus.Active && s.assigneeEmployeeId === selfId.value,
  onPreviewBumped: () => void loadSummary(true),
  onQueueChanged: () => void loadSummary(true),
})

const isQueuedActive = computed(() => active.value?.status === SupportSessionStatus.Queued)
const isClosedActive = computed(() => active.value?.status === SupportSessionStatus.Closed)
const isMineActive = computed(
  () => !!active.value && Number(active.value.assigneeEmployeeId) === Number(selfId.value),
)
const supportDeskPath = computed(
  () => findMenuPath(employeeStore.permissions, PERM.ImSupport) || '',
)
const showAccept = computed(() => isQueuedActive.value && canAccept.value)
const showTransferToSelf = computed(
  () =>
    active.value?.status === SupportSessionStatus.Active &&
    !isMineActive.value &&
    canTransfer.value,
)
const showPeerTransfer = computed(
  () => active.value?.status === SupportSessionStatus.Active && canTransfer.value,
)
const showGoDesk = computed(
  () => isMineActive.value && !isClosedActive.value && !!supportDeskPath.value,
)

const headerSub = computed(() => {
  const s = active.value
  if (!s) return ''
  const parts: string[] = []
  if (s.assigneeName) parts.push(`坐席 ${s.assigneeName}`)
  if (s.closeReasonText) parts.push(`结案原因：${s.closeReasonText}`)
  if (s.closeNote) parts.push(`备注：${s.closeNote}`)
  return parts.join(' · ')
})

const footerText = computed(() => {
  if (isQueuedActive.value) return '排队中会话只读预览，接入后才能回复'
  if (isClosedActive.value) return '会话已结案，仅可查看历史消息'
  if (isMineActive.value) return '当前坐席是你，请到「官方客服」工作台回复或结案'
  return '总览只读，消息每 5 秒自动刷新；如需介入请转接'
})

const listLoading = computed(() => {
  if (sideTab.value === 'queue') return queue.loading
  if (sideTab.value === 'active') return activeSessions.loading
  return closedSessions.loading
})

const tabs = computed(() => [
  {
    key: 'queue',
    label: '排队',
    count: summary.value.queueCount,
    warn: summary.value.longWaitQueueCount > 0,
  },
  { key: 'active', label: '进行中', count: summary.value.activeCount },
  { key: 'closed', label: '已结案' },
])

function firstQuery(v: LocationQueryValue | LocationQueryValue[] | undefined) {
  const x = Array.isArray(v) ? v[0] : v
  return typeof x === 'string' ? x : ''
}

function routeQueryKey() {
  const q = route.query
  const out: Record<string, string> = {}
  for (const key of ['tab', 'keyword', 'assignee', 'start', 'end', 'sid']) {
    const v = firstQuery(q[key])
    if (v) out[key] = v
  }
  return JSON.stringify(out)
}

function buildQuery() {
  const q: Record<string, string> = {}
  if (sideTab.value !== 'queue') q.tab = sideTab.value
  if (filters.keyword) q.keyword = filters.keyword
  if (sideTab.value !== 'queue') {
    if (filters.assigneeId) q.assignee = String(filters.assigneeId)
    if (filters.dateRange) {
      q.start = filters.dateRange[0]
      q.end = filters.dateRange[1]
    }
  }
  if (active.value) q.sid = String(active.value.id)
  return q
}

let lastQueryKey = ''

function syncUrl() {
  const q = buildQuery()
  const key = JSON.stringify(q)
  if (key === lastQueryKey && key === routeQueryKey()) return
  lastQueryKey = key
  void router.replace({ path: route.path, query: q })
}

/** 从地址栏恢复筛选条件，返回需要打开的会话 ID。 */
function applyRoute() {
  const q = route.query
  const tab = firstQuery(q.tab) as SideTab
  sideTab.value = SIDE_TABS.includes(tab) ? tab : 'queue'
  filters.keyword = firstQuery(q.keyword).trim()
  keywordDraft.value = filters.keyword
  const assignee = Number(firstQuery(q.assignee))
  filters.assigneeId = assignee > 0 ? assignee : undefined
  const startDate = firstQuery(q.start)
  const endDate = firstQuery(q.end)
  filters.dateRange = startDate && endDate ? [startDate, endDate] : null
  lastQueryKey = routeQueryKey()
  return Number(firstQuery(q.sid)) || 0
}

function resetActive() {
  active.value = null
  clearMessages()
}

function currentList() {
  if (sideTab.value === 'queue') return queue
  if (sideTab.value === 'active') return activeSessions
  return closedSessions
}

async function reloadCurrentTab() {
  await Promise.all([currentList().load(), loadSummary(true)])
}

function applyFilters() {
  filters.keyword = keywordDraft.value.trim()
  resetActive()
  syncUrl()
  void reloadCurrentTab()
}

function onTabChange(key: string) {
  const tab = key as SideTab
  if (tab === sideTab.value) return
  sideTab.value = tab
  resetActive()
  syncUrl()
  void reloadCurrentTab()
}

function softRefreshAll() {
  return Promise.all([
    loadSummary(true),
    queue.load({ silent: true }),
    activeSessions.load({ silent: true }),
    closedSessions.load({ silent: true }),
  ])
}

async function selectSession(s: SupportAdminSession) {
  active.value = s
  syncUrl()
  await loadMessages()
}

async function openSessionById(id: number) {
  try {
    const { data } = await getSupportSession(id)
    active.value = data
    await loadMessages()
  } catch (e) {
    showApiError(e, '会话不存在或无权查看')
    active.value = null
  }
}

/** 旁听时同步会话状态（例如排队会话被他人接入）。 */
async function refreshActiveSession() {
  const id = active.value?.id
  if (!id) return
  try {
    const { data } = await getSupportSession(id)
    if (active.value?.id === id) Object.assign(active.value, data)
  } catch {
    return
  }
}

async function goDeskOr(fallback: () => void | Promise<void>) {
  if (supportDeskPath.value) await router.push(supportDeskPath.value)
  else await fallback()
}

async function onAccept(s: SupportAdminSession) {
  if (!canAccept.value || accepting.value) return
  accepting.value = true
  try {
    await acceptSupport(s.id)
    ElMessage.success('已接入')
    await softRefreshAll()
    await goDeskOr(async () => {
      await refreshActiveSession()
      onTabChange('active')
    })
  } catch (e) {
    showApiError(e, '接入失败')
    await Promise.all([queue.load(), refreshActiveSession()])
  } finally {
    accepting.value = false
  }
}

async function onTransferToSelf() {
  const s = active.value
  if (!s || !showTransferToSelf.value || !selfId.value) return
  try {
    await ElMessageBox.confirm(
      `确认从坐席「${s.assigneeName || '未知'}」接管「${supportSessionTitle(s)}」？`,
      '转接给我',
    )
    await transferSupport(s.id, selfId.value)
    ElMessage.success('已接管')
    await softRefreshAll()
    await goDeskOr(refreshActiveSession)
  } catch (e) {
    if (isCancelAction(e)) return
    showApiError(e, '转接失败')
  }
}

async function onTransferred() {
  ElMessage.success('已转接')
  await Promise.all([softRefreshAll(), refreshActiveSession()])
}

function goSupportDesk() {
  if (supportDeskPath.value) void router.push(supportDeskPath.value)
}

async function initialize() {
  const sid = applyRoute()
  await Promise.all([
    loadSummary(),
    loadAgents(),
    queue.load(),
    activeSessions.load(),
    closedSessions.load(),
  ])
  if (sid) await openSessionById(sid)
  else if (!route.query.tab && summary.value.queueCount === 0 && summary.value.activeCount > 0) {
    sideTab.value = 'active'
  }
  syncUrl()
}

const pagePath = route.path

watch(
  () => route.query,
  () => {
    if (route.path !== pagePath || routeQueryKey() === lastQueryKey) return
    resetActive()
    void initialize()
  },
)

let livePollTimer = 0

onMounted(async () => {
  await initialize()
  start()
  livePollTimer = window.setInterval(() => {
    const s = active.value
    if (!s || s.status === SupportSessionStatus.Closed || !pageVisible.value) return
    void softRefreshMessages()
    void refreshActiveSession()
  }, LIVE_POLL_MS)
})

onUnmounted(() => window.clearInterval(livePollTimer))
</script>
