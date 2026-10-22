<template>
  <div class="admin-page im-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <ImAttentionToggle />
          </div>
        </div>
      </template>

      <ImConnectionBanner :state="wsState" @reconnect="reconnect" />

      <ImShell v-model:active-tab="sideTab" :tabs="tabs">
        <template #list>
          <div v-loading="listLoading">
            <SupportSessionList
              v-if="sideTab === 'queue'"
              :has-more="queue.hasMore"
              :loading-more="queue.loadingMore"
              :empty="queue.loaded && queue.list.length === 0"
              empty-text="暂无排队"
              @load-more="queue.loadMore"
            >
              <ImListItem
                v-for="s in queue.list"
                :key="s.id"
                :title="supportSessionTitle(s)"
                :preview="s.lastContent || '新会话'"
                :avatar-url="s.userAvatarUrl"
                :wait-text="`等待 ${imFormatWait(supportWaitSec(s, now))}`"
                :wait-tone="imWaitTone(supportWaitSec(s, now))"
              >
                <template v-if="canAccept" #action>
                  <el-button
                    size="small"
                    type="primary"
                    :loading="acceptingId === s.id"
                    @click="onAccept(s)"
                  >
                    接入
                  </el-button>
                </template>
              </ImListItem>
            </SupportSessionList>

            <SupportSessionList
              v-else-if="sideTab === 'mine'"
              :has-more="mine.hasMore"
              :loading-more="mine.loadingMore"
              :empty="mine.loaded && mine.list.length === 0"
              empty-text="暂无进行中会话"
              @load-more="mine.loadMore"
            >
              <ImListItem
                v-for="s in mine.list"
                :key="s.id"
                :title="supportSessionTitle(s)"
                :preview="s.lastContent || '暂无消息'"
                :time="imFormatTime(s.lastActiveAt)"
                :avatar-url="s.userAvatarUrl"
                :unread="s.unreadCount"
                :active="active?.id === s.id"
                @click="selectSession(s)"
              />
            </SupportSessionList>

            <SupportSessionList
              v-else
              :has-more="closed.hasMore"
              :loading-more="closed.loadingMore"
              :empty="closed.loaded && closed.list.length === 0"
              empty-text="暂无结案记录"
              @load-more="closed.loadMore"
            >
              <ImListItem
                v-for="s in closed.list"
                :key="s.id"
                :title="supportSessionTitle(s)"
                :preview="closedPreview(s)"
                :time="imFormatTime(s.closeAt || s.lastActiveAt)"
                :avatar-url="s.userAvatarUrl"
                :active="active?.id === s.id"
                @click="selectSession(s)"
              />
            </SupportSessionList>
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
                <el-tag v-if="isClosedActive" size="small" type="info" effect="plain">
                  已结束
                </el-tag>
              </template>
              <template v-if="isClosedActive && active.closeReasonText" #sub>
                结案原因：{{ active.closeReasonText }}
                <template v-if="active.closeNote"> · 备注：{{ active.closeNote }}</template>
              </template>
              <template #actions>
                <el-button
                  :type="historyVisible ? 'primary' : 'default'"
                  plain
                  @click="toggleHistory"
                >
                  历史咨询
                </el-button>
                <template v-if="!isClosedActive">
                  <el-button v-if="canTransfer" @click="transferVisible = true">转接</el-button>
                  <el-button v-if="canClose" type="danger" plain @click="closeVisible = true">
                    结案
                  </el-button>
                </template>
              </template>
            </ImChatHeader>
            <div class="support-desk__body">
              <div class="support-desk__chat">
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
                <ImComposer
                  v-if="!isClosedActive && canSend"
                  v-model="draft"
                  :sending="sending"
                  :uploading="uploading"
                  :allow-upload="canUpload"
                  :quick-replies="quickReplies"
                  placeholder="回复用户，Enter 发送；输入 / 选择快捷回复"
                  @send="onSend"
                  @upload="onUpload"
                />
                <div v-else-if="isClosedActive" class="im-closed-bar">
                  会话已结案，仅可查看历史消息
                </div>
                <div v-else class="im-closed-bar">无回复权限</div>
              </div>
              <SupportUserHistory v-if="historyVisible" :conversation-id="active.id" />
            </div>
          </template>
          <div v-else class="im-empty">
            <div class="im-empty__title">{{ route.meta.title }}</div>
            <div class="im-empty__desc">从排队池接入用户，或选择进行中 / 已结案会话</div>
            <el-button type="primary" @click="sideTab = 'queue'">查看排队</el-button>
          </div>
        </template>
      </ImShell>
    </el-card>

    <SupportTransferDialog
      v-model="transferVisible"
      title="转接会话"
      :conversation-id="active?.id"
      :exclude-ids="[selfId]"
      @transferred="onTransferred"
    />
    <SupportCloseDialog
      v-model="closeVisible"
      :conversation-id="active?.id"
      :session-title="active ? supportSessionTitle(active) : ''"
      @closed="onClosed"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useEmployeeStore } from '@/stores/employee'
import {
  acceptSupport,
  getSupportClosed,
  getSupportMine,
  getSupportQueue,
  getSupportQuickReplies,
  type SupportAdminSession,
  type SupportMessageItem,
  type SupportQuickReply,
} from '@/api/support'
import { PERM } from '@/constants/permission'
import {
  SUPPORT_CLOSE_REASON_OPTIONS,
  SUPPORT_LONG_WAIT_SEC,
  SupportSessionStatus,
} from '@/constants/support'
import { usePermission } from '@/composables/usePermission'
import { useSupportMessaging } from '@/composables/useSupportMessaging'
import {
  reselectActiveSession,
  supportSessionTitle,
  supportWaitSec,
  useNowSec,
  useSupportSessionList,
} from '@/composables/useSupportSessions'
import { showApiError } from '@/utils/apiError'
import { alertAttention, usePageVisible, useTitleCount } from '@/utils/attention'
import UserCell from '@/components/common/UserCell.vue'
import ImAttentionToggle from '@/components/im/ImAttentionToggle.vue'
import ImConnectionBanner from '@/components/im/ImConnectionBanner.vue'
import ImShell from '@/components/im/ImShell.vue'
import ImListItem from '@/components/im/ImListItem.vue'
import ImChatHeader from '@/components/im/ImChatHeader.vue'
import ImMessageThread from '@/components/im/ImMessageThread.vue'
import ImComposer from '@/components/im/ImComposer.vue'
import { imFormatTime, imFormatWait, imMessagePreview, imWaitTone } from '@/components/im/media'
import SupportCloseDialog from './SupportCloseDialog.vue'
import SupportSessionList from './SupportSessionList.vue'
import SupportTransferDialog from './SupportTransferDialog.vue'
import SupportUserHistory from './SupportUserHistory.vue'

const HISTORY_PREF_KEY = 'admin_support_desk_history'

const employeeStore = useEmployeeStore()
const { can } = usePermission()
const route = useRoute()
const now = useNowSec()
const pageVisible = usePageVisible()
const selfId = computed(() => employeeStore.employeeInfo?.id || 0)

const canAccept = computed(() => can(PERM.ImSupportAccept))
const canTransfer = computed(() => can(PERM.ImSupportTransfer))
const canClose = computed(() => can(PERM.ImSupportClose))
const canSend = computed(() => can(PERM.ImSupportSend))
const canUpload = computed(() => can(PERM.OpsUpload))

type SideTab = 'queue' | 'mine' | 'closed'

const sideTab = ref<SideTab>('queue')
const active = ref<SupportAdminSession | null>(null)
const transferVisible = ref(false)
const closeVisible = ref(false)
const acceptingId = ref(0)
const historyVisible = ref(localStorage.getItem(HISTORY_PREF_KEY) !== '0')
const quickReplies = ref<SupportQuickReply[]>([])
const knownQueueIds = new Set<number>()
let queueInitialized = false

const queue = useSupportSessionList({
  fetch: getSupportQueue,
  errorText: '加载排队失败',
  onFetched: detectNewQueueItems,
})
const mine = useSupportSessionList({
  fetch: getSupportMine,
  errorText: '加载会话失败',
  onReplaced: (list) =>
    reselectActiveSession(active, list, (s) => s.status !== SupportSessionStatus.Closed),
})
const closed = useSupportSessionList({
  fetch: getSupportClosed,
  errorText: '加载结案记录失败',
  onReplaced: (list) =>
    reselectActiveSession(
      active,
      list,
      (s) => s.status === SupportSessionStatus.Closed && sideTab.value === 'closed',
    ),
})

const {
  messages,
  messagesLoading,
  messagesLoadingMore,
  messagesHasMore,
  threadRef,
  draft,
  sending,
  uploading,
  wsState,
  reconnect,
  start,
  loadMessages,
  loadMoreMessages,
  clearMessages,
  onSend,
  onUpload,
} = useSupportMessaging(active, {
  unreadList: mine,
  previewLists: [queue],
  refreshLists: softRefreshLists,
  canSend: (s) => s.status !== SupportSessionStatus.Closed && canSend.value,
  canUpload: () => canUpload.value,
  canMarkRead: (s) =>
    s.status === SupportSessionStatus.Active && s.assigneeEmployeeId === selfId.value,
  onIncoming: onIncomingMessage,
})

const isClosedActive = computed(() => active.value?.status === SupportSessionStatus.Closed)
const listLoading = computed(() => {
  if (sideTab.value === 'queue') return queue.loading
  if (sideTab.value === 'mine') return mine.loading
  return closed.loading
})
const mineUnread = computed(() => mine.list.reduce((n, s) => n + (s.unreadCount || 0), 0))

const tabs = computed(() => [
  {
    key: 'queue',
    label: '排队',
    count: queue.total,
    warn: queue.list.some((s) => supportWaitSec(s, now.value) >= SUPPORT_LONG_WAIT_SEC),
  },
  { key: 'mine', label: '进行中', count: mine.total, unread: mineUnread.value },
  { key: 'closed', label: '已结案' },
])

useTitleCount('support-desk', () => mineUnread.value + queue.total)

function closedPreview(s: SupportAdminSession) {
  const reason = s.closeReasonText ? `[${s.closeReasonText}] ` : ''
  return `${reason}${s.lastContent || '暂无消息'}`
}

function detectNewQueueItems(page: SupportAdminSession[]) {
  const fresh = page.filter((s) => !knownQueueIds.has(s.id))
  page.forEach((s) => knownQueueIds.add(s.id))
  if (!queueInitialized) {
    queueInitialized = true
    return
  }
  const first = fresh[0]
  if (!first || !canAccept.value) return
  const more = fresh.length > 1 ? ` 等 ${fresh.length} 位用户` : ''
  alertAttention(
    '新的排队咨询',
    `${supportSessionTitle(first)}${more}正在等待接入`,
    'support-queue',
  )
}

function onIncomingMessage(item: SupportMessageItem) {
  const convId = Number(item.conversationId)
  if (pageVisible.value && active.value?.id === convId) return
  const s = mine.list.find((x) => x.id === convId)
  const name = s ? supportSessionTitle(s) : item.senderName || '用户'
  alertAttention(`${name} 发来新消息`, imMessagePreview(item), `support-${convId}`)
}

function softRefreshLists() {
  return Promise.all([
    queue.load({ silent: true }),
    mine.load({ silent: true }),
    closed.load({ silent: true }),
  ])
}

function toggleHistory() {
  historyVisible.value = !historyVisible.value
  localStorage.setItem(HISTORY_PREF_KEY, historyVisible.value ? '1' : '0')
}

async function onAccept(s: SupportAdminSession) {
  if (!canAccept.value || acceptingId.value) return
  acceptingId.value = s.id
  try {
    const { data } = await acceptSupport(s.id)
    sideTab.value = 'mine'
    active.value = data
    await Promise.all([queue.load(), mine.load(), loadMessages()])
  } catch (e) {
    showApiError(e, '接入失败')
    await queue.load()
  } finally {
    acceptingId.value = 0
  }
}

async function selectSession(s: SupportAdminSession) {
  active.value = s
  await loadMessages()
}

async function onClosed(reason: string, note: string) {
  if (!active.value) return
  ElMessage.success('已结案')
  active.value = {
    ...active.value,
    status: SupportSessionStatus.Closed,
    statusText: '已结束',
    unreadCount: 0,
    closeAt: Math.floor(Date.now() / 1000),
    closeReason: reason,
    closeReasonText: SUPPORT_CLOSE_REASON_OPTIONS.find((o) => o.value === reason)?.label || '',
    closeNote: note,
  }
  sideTab.value = 'closed'
  await Promise.all([mine.load(), closed.load(), loadMessages()])
}

async function onTransferred() {
  active.value = null
  clearMessages()
  await mine.load()
  ElMessage.success('已转接')
}

async function loadQuickReplies() {
  if (!canSend.value) return
  try {
    const { data } = await getSupportQuickReplies()
    quickReplies.value = data.list || []
  } catch {
    quickReplies.value = []
  }
}

onMounted(async () => {
  await Promise.all([queue.load(), mine.load(), closed.load(), loadQuickReplies()])
  if (queue.list.length === 0 && mine.list.length > 0) sideTab.value = 'mine'
  start()
})
</script>

<style scoped>
.support-desk__body {
  flex: 1;
  min-height: 0;
  display: flex;
}

.support-desk__chat {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
</style>
