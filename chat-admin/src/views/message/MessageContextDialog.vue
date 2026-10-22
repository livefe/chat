<template>
  <el-dialog v-model="visible" :title="title" width="760px" destroy-on-close>
    <div ref="scrollRef" v-loading="loading" class="context-list">
      <div v-if="hasMoreBefore && list.length" class="context-more">
        <el-button link type="primary" :loading="loadingBefore" @click="loadBefore">
          加载更早
        </el-button>
      </div>
      <div
        v-for="item in list"
        :key="item.id"
        :data-id="item.id"
        class="context-item"
        :class="{ highlight: item.id === centerId }"
      >
        <div class="context-meta">
          <UserCell :user-id="item.fromUserId" :nickname="item.fromName" />
          <template v-if="item.scope === MessageScope.Private">
            <span class="sub">→</span>
            <UserCell :user-id="item.toUserId" :nickname="item.toName" />
          </template>
          <span class="time">{{ formatUnixSecondsLocale(item.createdAt) }}</span>
          <el-tag v-if="item.deleted" type="danger" size="small">已删除</el-tag>
          <el-tag v-if="item.recalled" type="warning" size="small">已撤回</el-tag>
          <span class="sub context-id">#{{ item.id }}</span>
        </div>
        <AdminReplyQuote :item="item" />
        <div class="context-content">
          <MessageContent :item="item" :muted="item.deleted || item.recalled" />
        </div>
      </div>
      <div v-if="hasMoreAfter && list.length" class="context-more">
        <el-button link type="primary" :loading="loadingAfter" @click="loadAfter">
          加载更晚
        </el-button>
      </div>
      <el-empty v-if="!loading && list.length === 0" description="暂无上下文" />
    </div>
    <template v-if="center" #footer>
      <div class="dialog-footer">
        <span v-if="center.scope === MessageScope.Nearby" class="sub context-hint">
          附近消息按发送位置周边 3 公里展示
        </span>
        <el-button v-if="conversationQuery" @click="goConversation">查看完整会话</el-button>
        <el-button
          v-if="center.scope === MessageScope.Room && center.roomId && can(PERM.ChatRoom)"
          @click="go(PERM.ChatRoom, { id: String(center.roomId) })"
        >
          查看聊天区
        </el-button>
        <el-button
          v-if="center.scope === MessageScope.Post && center.postId && can(PERM.PostPost)"
          @click="go(PERM.PostPost, { id: String(center.postId) })"
        >
          查看帖子
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getMessageContext, type AdminMessageItem } from '@/api/message'
import AdminReplyQuote from '@/components/AdminReplyQuote.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { MessageScope } from '@/constants/message'
import { PERM, type PermissionCode } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { formatUnixSecondsLocale } from '@/utils/datetime'
import { pushMenuRoute } from '@/utils/navigation'
import MessageContent from './MessageContent.vue'

const PAGE = 20

const router = useRouter()
const employeeStore = useEmployeeStore()
const { can } = usePermission()

const visible = ref(false)
const loading = ref(false)
const loadingBefore = ref(false)
const loadingAfter = ref(false)
const list = ref<AdminMessageItem[]>([])
const centerId = ref(0)
const hasMoreBefore = ref(false)
const hasMoreAfter = ref(false)
const scrollRef = ref<HTMLElement>()

const center = computed(() => list.value.find((m) => m.id === centerId.value))
const title = computed(() => {
  const c = center.value
  if (!c) return '消息上下文'
  if (c.scope === MessageScope.Room) return `消息上下文 · ${c.roomName || '聊天区'}`
  return `消息上下文 · ${c.scopeText}`
})

const conversationQuery = computed((): Record<string, string> | null => {
  const c = center.value
  if (!c) return null
  if (c.scope === MessageScope.Private) {
    return {
      userId: String(c.fromUserId),
      peerUserId: String(c.toUserId),
      scope: String(MessageScope.Private),
    }
  }
  if (c.scope === MessageScope.Room && c.roomId) {
    return { roomId: String(c.roomId), scope: String(MessageScope.Room) }
  }
  if (c.scope === MessageScope.Post && c.postId) {
    return { postId: String(c.postId), scope: String(MessageScope.Post) }
  }
  return null
})

async function open(id: number) {
  centerId.value = id
  visible.value = true
  loading.value = true
  list.value = []
  try {
    const { data } = await getMessageContext(id, { limit: PAGE })
    list.value = data.list ?? []
    const idx = list.value.findIndex((m) => m.id === id)
    hasMoreBefore.value = idx >= PAGE
    hasMoreAfter.value = idx >= 0 && list.value.length - idx - 1 >= PAGE
    await nextTick()
    scrollRef.value?.querySelector(`[data-id="${id}"]`)?.scrollIntoView({ block: 'center' })
  } catch (e) {
    showApiError(e, '获取上下文失败')
  } finally {
    loading.value = false
  }
}

async function loadBefore() {
  const first = list.value[0]
  if (!first) return
  loadingBefore.value = true
  try {
    const { data } = await getMessageContext(centerId.value, { limit: PAGE, beforeId: first.id })
    const rows = data.list ?? []
    const el = scrollRef.value
    const prevHeight = el?.scrollHeight ?? 0
    list.value = [...rows, ...list.value]
    hasMoreBefore.value = rows.length >= PAGE
    await nextTick()
    if (el) el.scrollTop += el.scrollHeight - prevHeight
  } catch (e) {
    showApiError(e, '加载失败')
  } finally {
    loadingBefore.value = false
  }
}

async function loadAfter() {
  const last = list.value[list.value.length - 1]
  if (!last) return
  loadingAfter.value = true
  try {
    const { data } = await getMessageContext(centerId.value, { limit: PAGE, afterId: last.id })
    const rows = data.list ?? []
    list.value = [...list.value, ...rows]
    hasMoreAfter.value = rows.length >= PAGE
  } catch (e) {
    showApiError(e, '加载失败')
  } finally {
    loadingAfter.value = false
  }
}

function go(code: PermissionCode, query: Record<string, string>) {
  visible.value = false
  pushMenuRoute(router, employeeStore.permissions, code, query)
}

function goConversation() {
  if (conversationQuery.value) go(PERM.ChatMessage, conversationQuery.value)
}

defineExpose({ open })
</script>

<style scoped>
.context-list {
  max-height: 520px;
  min-height: 120px;
  overflow-y: auto;
}

.context-more {
  padding: 6px 0;
  text-align: center;
}

.context-item {
  padding: 10px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.context-item.highlight {
  background: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.context-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}

.context-meta .time {
  color: var(--el-text-color-secondary);
}

.context-id {
  margin-left: auto;
}

.context-content {
  color: var(--el-text-color-primary);
  line-height: 1.5;
  word-break: break-word;
}

.context-hint {
  margin-right: auto;
}
</style>
