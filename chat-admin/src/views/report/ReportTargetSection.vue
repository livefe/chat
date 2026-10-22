<template>
  <section class="review-section">
    <div class="section-head">
      <div class="section-title">被举报内容</div>
      <div class="section-actions">
        <el-button
          v-if="report.targetUserId && can(PERM.UserAccountDetail)"
          size="small"
          link
          type="primary"
          @click="drawer.open(report.targetUserId)"
        >
          查看用户
        </el-button>
        <el-button
          v-if="postId && can(PERM.PostPost)"
          size="small"
          link
          type="primary"
          @click="go(PERM.PostPost, { id: String(postId) })"
        >
          查看帖子
        </el-button>
        <el-button
          v-if="report.targetType === ReportTargetType.Message && can(PERM.ChatMessage)"
          size="small"
          link
          type="primary"
          @click="go(PERM.ChatMessage, { contextId: String(report.targetId) })"
        >
          查看消息
        </el-button>
      </div>
    </div>

    <div class="target-card">
      <template v-if="report.targetType === ReportTargetType.User">
        <div class="target-name">
          {{ report.context.nickname || report.targetUserName || '-' }}
        </div>
        <div class="target-meta">用户 ID {{ report.targetId }}（昵称为举报时快照）</div>
      </template>

      <template v-else-if="report.targetType === ReportTargetType.Post">
        <div class="target-name">
          {{ report.context.postTitle || `帖子 #${report.targetId}` }}
        </div>
        <div v-if="report.context.postContent" class="target-content">
          {{ report.context.postContent }}
        </div>
        <div v-if="postImages.length" class="images">
          <el-image
            v-for="(url, i) in postImages"
            :key="i"
            :src="url"
            :preview-src-list="postImages"
            fit="cover"
            class="thumb"
          />
        </div>
      </template>

      <template v-else-if="report.targetType === ReportTargetType.Message">
        <div class="target-meta">
          {{ messageScopeLabel(report.context.scope) }}
          <template v-if="report.context.scope === MessageScope.Room && report.context.roomId">
            · 聊天区 ID {{ report.context.roomId }}
          </template>
          <template v-if="report.context.sentAt"> · {{ report.context.sentAt }}</template>
        </div>
        <div
          v-if="report.context.scope === MessageScope.Private && report.context.peerId"
          class="peer"
        >
          <span class="sub">私聊对象</span>
          <UserCell :user-id="report.context.peerId" />
        </div>
        <div class="target-content highlight">{{ messageText || '-' }}</div>
        <div v-if="messageContext.length" class="message-context">
          <div class="context-label">消息上下文</div>
          <div
            v-for="item in messageContext"
            :key="item.id"
            class="context-item"
            :class="{ 'is-target': item.id === report.targetId }"
          >
            <div class="context-meta">
              <span class="context-user">
                <UserCell :user-id="item.fromUserId" :nickname="item.fromName" />
                <el-tag v-if="item.recalled" type="warning" size="small">已撤回</el-tag>
                <el-tag v-if="item.deleted" type="danger" size="small">已删除</el-tag>
              </span>
              <span>{{ formatUnixSecondsShort(item.createdAt) }}</span>
            </div>
            <AdminReplyQuote :item="item" />
            <div class="context-content" :class="{ 'is-muted': item.recalled || item.deleted }">
              {{ formatMessageSummary(item.contentType, item.content) }}
            </div>
          </div>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import type { AdminMessageItem } from '@/api/message'
import type { ReportItem } from '@/api/report'
import AdminReplyQuote from '@/components/AdminReplyQuote.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { MessageContentType, MessageScope } from '@/constants/message'
import { PERM, type PermissionCode } from '@/constants/permission'
import { ReportTargetType } from '@/constants/report'
import { useEmployeeStore } from '@/stores/employee'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { formatUnixSecondsShort } from '@/utils/datetime'
import { formatMessageSummary } from '@/utils/messageDisplay'
import { pushMenuRoute } from '@/utils/navigation'
import { messageScopeLabel } from '@/utils/reportDisplay'

const props = defineProps<{
  report: ReportItem
  messageContext: AdminMessageItem[]
}>()
const emit = defineEmits<{ leave: [] }>()

const router = useRouter()
const employeeStore = useEmployeeStore()
const drawer = useUserDrawerStore()
const { can } = usePermission()

const postImages = computed(() => props.report.context.postImages ?? [])
const messageText = computed(() =>
  formatMessageSummary(
    props.report.context.messageType ?? MessageContentType.Text,
    props.report.context.messageContent ?? '',
  ),
)
const postId = computed(() => {
  if (props.report.targetType === ReportTargetType.Post) return props.report.targetId
  if (props.report.context.scope === MessageScope.Post) return props.report.context.postId ?? 0
  return 0
})

function go(code: PermissionCode, query: Record<string, string>) {
  emit('leave')
  pushMenuRoute(router, employeeStore.permissions, code, query)
}
</script>

<style scoped>
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.section-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.target-card {
  padding: 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-fill-color-light);
}

.target-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
}

.target-meta {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.peer {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.target-content {
  font-size: 14px;
  line-height: 1.7;
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.target-content.highlight {
  padding: 10px 12px;
  border-radius: 6px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
}

.context-label {
  width: 100%;
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.message-context {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px dashed var(--el-border-color);
}

.context-item {
  padding: 8px 10px;
  border-radius: 6px;
  margin-bottom: 8px;
  background: var(--el-bg-color);
}

.context-item.is-target {
  border: 1px solid var(--el-color-danger);
  background: var(--el-color-danger-light-9);
}

.context-meta {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.context-user {
  display: flex;
  align-items: center;
  gap: 6px;
}

.context-content {
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.context-content.is-muted {
  color: var(--el-text-color-secondary);
  text-decoration: line-through;
}
</style>
