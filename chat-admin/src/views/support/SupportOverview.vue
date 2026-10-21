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

      <ImShell v-model:active-tab="sideTab" :tabs="tabs">
        <template #search>
          <SupportOverviewSearch
            v-model:keyword="keywordDraft"
            v-model:assignee-id="assigneeId"
            :show-assignee="sideTab !== 'queue'"
            :agent-options="agentOptions"
            @search="applyFilters"
            @change="onFilterChange"
          />
        </template>

        <template #list>
          <div v-loading="listLoading">
            <SupportOverviewSessionLists
              :tab="sideTab"
              :queue="queue"
              :queue-items="filteredQueue"
              :active-sessions="activeSessions"
              :closed-sessions="closedSessions"
              :active-id="active?.id"
              :can-accept="canAccept"
              @select="selectSession"
              @accept="onAccept"
            />
          </div>
        </template>

        <template #main>
          <template v-if="active">
            <ImChatHeader :title="supportSessionTitle(active)">
              <template #tags>
                <el-tag v-if="isQueuedActive" size="small" type="warning" effect="plain"
                  >排队中</el-tag
                >
                <el-tag v-else-if="isClosedActive" size="small" type="info" effect="plain"
                  >已结束</el-tag
                >
              </template>
              <template #sub>
                UID {{ active.userId }}
                <template v-if="active.userMobile"> · {{ active.userMobile }}</template>
                <template v-if="active.assigneeName"> · 坐席 {{ active.assigneeName }}</template>
              </template>
              <template #actions>
                <el-button v-if="showAccept" type="primary" @click="onAccept(active)"
                  >接入</el-button
                >
                <el-button v-if="showTransferToSelf" type="primary" @click="onTransferToSelf">
                  转接给我
                </el-button>
                <el-button v-if="showPeerTransfer" @click="openTransfer">转接给同事</el-button>
                <el-button v-if="showGoDesk" @click="goSupportDesk">去工作台</el-button>
              </template>
            </ImChatHeader>
            <ImMessageThread
              v-if="!isQueuedActive"
              ref="threadRef"
              :messages="messages"
              :self-id="selfId"
              mine-mode="support"
              :loading="messagesLoading"
              :has-more="messagesHasMore"
              :loading-more="messagesLoadingMore"
              @load-more="loadMoreMessages"
            />
            <div v-if="isQueuedActive" class="im-closed-bar">排队中会话接入后可查看消息</div>
            <div v-else-if="isClosedActive" class="im-closed-bar">会话已结案，仅可查看历史消息</div>
            <div v-else-if="isMineActive" class="im-closed-bar">
              当前坐席是你，请到「官方客服」工作台回复或结案
            </div>
            <div v-else class="im-closed-bar">总览只读，如需介入请转接</div>
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

    <el-drawer v-model="configVisible" title="客服配置" size="480px" destroy-on-close>
      <SupportConfig />
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Setting } from '@element-plus/icons-vue'
import { useEmployeeStore } from '@/stores/employee'
import {
  acceptSupport,
  getSupportActiveAll,
  getSupportClosedAll,
  getSupportQueue,
  transferSupport,
  type SupportAdminSession,
} from '@/api/support'
import { PERM } from '@/constants/permission'
import { SupportSessionStatus } from '@/constants/support'
import { usePermission } from '@/composables/usePermission'
import { useSupportMessaging } from '@/composables/useSupportMessaging'
import {
  useSupportOverviewFilters,
  useSupportOverviewSummary,
} from '@/composables/useSupportOverview'
import {
  reselectActiveSession,
  supportSessionTitle,
  useSupportSessionList,
} from '@/composables/useSupportSessions'
import { showApiError } from '@/utils/apiError'
import { isCancelAction } from '@/utils/confirm'
import { findMenuPath } from '@/utils/navigation'
import ImConnectionBanner from '@/components/im/ImConnectionBanner.vue'
import ImShell from '@/components/im/ImShell.vue'
import ImChatHeader from '@/components/im/ImChatHeader.vue'
import ImMessageThread from '@/components/im/ImMessageThread.vue'
import SupportConfig from './SupportConfig.vue'
import SupportOverviewSearch from './SupportOverviewSearch.vue'
import SupportOverviewSessionLists from './SupportOverviewSessionLists.vue'
import SupportTransferDialog from './SupportTransferDialog.vue'

const employeeStore = useEmployeeStore()
const { can } = usePermission()
const route = useRoute()
const router = useRouter()
const selfId = computed(() => employeeStore.employeeInfo?.id || 0)

const canAccept = computed(() => can(PERM.ImSupportAccept))
const canTransfer = computed(() => can(PERM.ImSupportTransfer))
const canOpenConfig = computed(() => can(PERM.ImSupportOverviewConfig))

type SideTab = 'queue' | 'active' | 'closed'

const configVisible = ref(false)
const sideTab = ref<SideTab>('queue')
const active = ref<SupportAdminSession | null>(null)
const transferVisible = ref(false)

const { summary, loadSummary } = useSupportOverviewSummary()
const {
  keywordDraft,
  assigneeId,
  agentOptions,
  applyKeyword,
  listFilterParams,
  filterQueue,
  mergeAgentOptions,
  loadAgents,
} = useSupportOverviewFilters()

const queue = useSupportSessionList({
  fetch: getSupportQueue,
  errorText: '加载排队失败',
  mode: 'replace',
})
const activeSessions = useSupportSessionList({
  fetch: (p) => getSupportActiveAll({ ...p, ...listFilterParams() }),
  errorText: '加载进行中失败',
  mode: 'replace',
  onReplaced: (list) =>
    reselectActiveSession(
      active,
      list,
      (s) => s.status === SupportSessionStatus.Active && sideTab.value === 'active',
    ),
  onFetched: mergeAgentOptions,
})
const closedSessions = useSupportSessionList({
  fetch: (p) => getSupportClosedAll({ ...p, ...listFilterParams() }),
  errorText: '加载已结案失败',
  mode: 'replace',
  onReplaced: (list) =>
    reselectActiveSession(
      active,
      list,
      (s) => s.status === SupportSessionStatus.Closed && sideTab.value === 'closed',
    ),
  onFetched: mergeAgentOptions,
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
  clearMessages,
} = useSupportMessaging(active, {
  unreadList: activeSessions,
  previewLists: [queue],
  refreshLists: softRefreshAll,
  canLoad: (s) => s.status !== SupportSessionStatus.Queued,
  onPreviewBumped: () => void loadSummary(true),
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
  () => !isQueuedActive.value && !isClosedActive.value && !isMineActive.value && canTransfer.value,
)
const showPeerTransfer = computed(
  () => !isQueuedActive.value && !isClosedActive.value && canTransfer.value,
)
const showGoDesk = computed(
  () => isMineActive.value && !isClosedActive.value && !!supportDeskPath.value,
)

const listLoading = computed(() => {
  if (sideTab.value === 'queue') return queue.loading
  if (sideTab.value === 'active') return activeSessions.loading
  return closedSessions.loading
})

const filteredQueue = computed(() => filterQueue(queue.list))

const tabs = computed(() => [
  {
    key: 'queue',
    label: '排队',
    count: summary.value.queueCount || undefined,
    warn: summary.value.longWaitQueueCount > 0,
  },
  { key: 'active', label: '进行中', count: summary.value.activeCount || undefined },
  { key: 'closed', label: '已结案', count: closedSessions.total || undefined },
])

function resetActive() {
  active.value = null
  clearMessages()
}

function applyFilters() {
  applyKeyword()
  resetActive()
  void reloadCurrentTab()
}

function onFilterChange() {
  resetActive()
  void reloadCurrentTab()
}

async function reloadCurrentTab() {
  if (sideTab.value === 'queue') await queue.load()
  else if (sideTab.value === 'active') await activeSessions.load()
  else await closedSessions.load()
  await loadSummary(true)
}

function softRefreshAll() {
  return Promise.all([
    loadSummary(true),
    queue.load({ silent: true }),
    activeSessions.load({ silent: true }),
    closedSessions.load({ silent: true }),
  ])
}

watch(sideTab, () => {
  resetActive()
  void reloadCurrentTab()
})

async function selectSession(s: SupportAdminSession) {
  active.value = s
  if (s.status === SupportSessionStatus.Queued) {
    clearMessages()
    return
  }
  await loadMessages()
}

async function goDeskOr(fallback: () => void | Promise<void>) {
  if (supportDeskPath.value) await router.push(supportDeskPath.value)
  else await fallback()
}

async function onAccept(s: SupportAdminSession) {
  if (!canAccept.value) return
  try {
    await ElMessageBox.confirm(`确认接入「${supportSessionTitle(s)}」？`, '接入会话')
    await acceptSupport(s.id)
    ElMessage.success('已接入')
    await softRefreshAll()
    await goDeskOr(async () => {
      sideTab.value = 'active'
      await activeSessions.load()
    })
  } catch (e) {
    if (isCancelAction(e)) return
    showApiError(e, '接入失败')
    await queue.load()
  }
}

async function onTransferToSelf() {
  if (!active.value || !showTransferToSelf.value || !selfId.value) return
  try {
    await ElMessageBox.confirm('确认将本会话转接给自己并接管？', '转接给我')
    await transferSupport(active.value.id, selfId.value)
    ElMessage.success('已接管')
    await softRefreshAll()
    await goDeskOr(() => {
      if (!active.value) return
      active.value = {
        ...active.value,
        assigneeEmployeeId: selfId.value,
        assigneeName: employeeStore.employeeInfo?.name || active.value.assigneeName,
      }
    })
  } catch (e) {
    if (isCancelAction(e)) return
    showApiError(e, '转接失败')
  }
}

function openTransfer() {
  if (!active.value || !showPeerTransfer.value) return
  transferVisible.value = true
}

async function onTransferred() {
  resetActive()
  await softRefreshAll()
  ElMessage.success('已转接')
}

function goSupportDesk() {
  if (supportDeskPath.value) void router.push(supportDeskPath.value)
}

onMounted(async () => {
  await Promise.all([
    loadSummary(),
    loadAgents(),
    queue.load(),
    activeSessions.load(),
    closedSessions.load(),
  ])
  if (summary.value.queueCount === 0 && summary.value.activeCount > 0) sideTab.value = 'active'
  start()
})
</script>
