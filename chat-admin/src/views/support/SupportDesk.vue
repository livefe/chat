<template>
  <div class="admin-page im-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
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
              :empty="!queue.loading && queue.list.length === 0"
              empty-text="暂无排队"
              @load-more="queue.loadMore"
            >
              <ImListItem
                v-for="s in queue.list"
                :key="s.id"
                :title="supportSessionTitle(s)"
                :preview="s.lastContent || '新会话'"
                :avatar-url="s.userAvatarUrl"
                :wait-text="`等待 ${imFormatWait(s.queueWaitSec)}`"
                :wait-tone="imWaitTone(s.queueWaitSec)"
                @click="canAccept ? onAccept(s) : undefined"
              >
                <template v-if="canAccept" #action>
                  <el-button size="small" type="primary" @click.stop="onAccept(s)">接入</el-button>
                </template>
              </ImListItem>
            </SupportSessionList>

            <SupportSessionList
              v-else-if="sideTab === 'mine'"
              :has-more="mine.hasMore"
              :loading-more="mine.loadingMore"
              :empty="!mine.loading && mine.list.length === 0"
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
              v-else-if="sideTab === 'closed'"
              :has-more="closed.hasMore"
              :loading-more="closed.loadingMore"
              :empty="!closed.loading && closed.list.length === 0"
              empty-text="暂无结案记录"
              @load-more="closed.loadMore"
            >
              <ImListItem
                v-for="s in closed.list"
                :key="s.id"
                :title="supportSessionTitle(s)"
                :preview="s.lastContent || '暂无消息'"
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
            <ImChatHeader :title="supportSessionTitle(active)">
              <template #tags>
                <el-tag v-if="isClosedActive" size="small" type="info" effect="plain"
                  >已结束</el-tag
                >
              </template>
              <template #sub>
                UID {{ active.userId }}
                <template v-if="active.userMobile"> · {{ active.userMobile }}</template>
              </template>
              <template v-if="!isClosedActive" #actions>
                <el-button v-if="canTransfer" @click="openTransfer">转接</el-button>
                <el-button v-if="canClose" type="danger" plain @click="onClose">结案</el-button>
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
            <ImComposer
              v-if="!isClosedActive && canSend"
              v-model="draft"
              :sending="sending"
              :uploading="uploading"
              :allow-upload="canUpload"
              placeholder="回复用户，Enter 发送"
              @send="onSend"
              @upload="onUpload"
            />
            <div v-else-if="isClosedActive" class="im-closed-bar">会话已结案，仅可查看历史消息</div>
            <div v-else class="im-closed-bar">无回复权限</div>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useEmployeeStore } from '@/stores/employee'
import {
  acceptSupport,
  closeSupport,
  getSupportClosed,
  getSupportMine,
  getSupportQueue,
  type SupportAdminSession,
} from '@/api/support'
import { PERM } from '@/constants/permission'
import { SupportSessionStatus } from '@/constants/support'
import { usePermission } from '@/composables/usePermission'
import { useSupportMessaging } from '@/composables/useSupportMessaging'
import {
  reselectActiveSession,
  supportSessionTitle,
  useSupportSessionList,
} from '@/composables/useSupportSessions'
import { showApiError } from '@/utils/apiError'
import { isCancelAction } from '@/utils/confirm'
import ImConnectionBanner from '@/components/im/ImConnectionBanner.vue'
import ImShell from '@/components/im/ImShell.vue'
import ImListItem from '@/components/im/ImListItem.vue'
import ImChatHeader from '@/components/im/ImChatHeader.vue'
import ImMessageThread from '@/components/im/ImMessageThread.vue'
import ImComposer from '@/components/im/ImComposer.vue'
import { imFormatTime, imFormatWait, imWaitTone } from '@/components/im/media'
import SupportSessionList from './SupportSessionList.vue'
import SupportTransferDialog from './SupportTransferDialog.vue'

const employeeStore = useEmployeeStore()
const { can } = usePermission()
const route = useRoute()
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

const queue = useSupportSessionList({
  fetch: getSupportQueue,
  errorText: '加载排队失败',
  mode: 'merge',
})
const mine = useSupportSessionList({
  fetch: getSupportMine,
  errorText: '加载会话失败',
  mode: 'merge',
  onReplaced: (list) =>
    reselectActiveSession(active, list, (s) => s.status !== SupportSessionStatus.Closed),
})
const closed = useSupportSessionList({
  fetch: getSupportClosed,
  errorText: '加载结案记录失败',
  mode: 'merge',
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
})

const isClosedActive = computed(() => active.value?.status === SupportSessionStatus.Closed)
const listLoading = computed(() => {
  if (sideTab.value === 'queue') return queue.loading
  if (sideTab.value === 'mine') return mine.loading
  return closed.loading
})

const tabs = computed(() => [
  {
    key: 'queue',
    label: '排队',
    count: queue.total || undefined,
    warn: queue.list.some((s) => (s.queueWaitSec || 0) >= 120),
  },
  {
    key: 'mine',
    label: '进行中',
    count: mine.list.reduce((n, s) => n + (s.unreadCount || 0), 0) || undefined,
  },
  { key: 'closed', label: '已结案', count: closed.total || undefined },
])

function softRefreshLists() {
  return Promise.all([
    queue.load({ silent: true }),
    mine.load({ silent: true }),
    closed.load({ silent: true }),
  ])
}

async function onAccept(s: SupportAdminSession) {
  if (!canAccept.value) return
  try {
    await ElMessageBox.confirm(`确认接入「${supportSessionTitle(s)}」？`, '接入会话')
    const { data } = await acceptSupport(s.id)
    await Promise.all([queue.load(), mine.load()])
    sideTab.value = 'mine'
    active.value = data
    await loadMessages()
  } catch (e) {
    if (isCancelAction(e)) return
    showApiError(e, '接入失败')
    await queue.load()
  }
}

async function selectSession(s: SupportAdminSession) {
  active.value = s
  await loadMessages()
}

async function onClose() {
  if (!active.value || !canClose.value) return
  try {
    await ElMessageBox.confirm('确认结案该客服会话？', '结案')
    await closeSupport(active.value.id)
    active.value = {
      ...active.value,
      status: SupportSessionStatus.Closed,
      statusText: '已结束',
      closeAt: Math.floor(Date.now() / 1000),
    }
    sideTab.value = 'closed'
    await Promise.all([mine.load(), closed.load()])
    await loadMessages()
    ElMessage.success('已结案')
  } catch (e) {
    if (isCancelAction(e)) return
    showApiError(e, '结案失败')
  }
}

function openTransfer() {
  if (!active.value || !canTransfer.value) return
  transferVisible.value = true
}

async function onTransferred() {
  active.value = null
  clearMessages()
  await mine.load()
  ElMessage.success('已转接')
}

onMounted(async () => {
  await softRefreshLists()
  if (queue.list.length === 0 && mine.list.length > 0) sideTab.value = 'mine'
  start()
})
</script>
