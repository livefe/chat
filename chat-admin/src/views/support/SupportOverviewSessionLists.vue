<template>
  <SupportSessionList
    v-if="tab === 'queue'"
    :has-more="queue.hasMore"
    :loading-more="queue.loadingMore"
    :empty="!queue.loading && queueItems.length === 0"
    empty-text="暂无排队"
    @load-more="queue.loadMore"
  >
    <ImListItem
      v-for="s in queueItems"
      :key="s.id"
      :title="supportSessionTitle(s)"
      :preview="s.lastContent || '新会话'"
      :avatar-url="s.userAvatarUrl"
      :wait-text="`等待 ${imFormatWait(s.queueWaitSec)}`"
      :wait-tone="imWaitTone(s.queueWaitSec)"
      @click="canAccept ? emit('accept', s) : emit('select', s)"
    >
      <template v-if="canAccept" #action>
        <el-button size="small" type="primary" @click.stop="emit('accept', s)">接入</el-button>
      </template>
    </ImListItem>
  </SupportSessionList>

  <SupportSessionList
    v-else-if="tab === 'active'"
    :has-more="activeSessions.hasMore"
    :loading-more="activeSessions.loadingMore"
    :empty="!activeSessions.loading && activeSessions.list.length === 0"
    empty-text="暂无进行中会话"
    @load-more="activeSessions.loadMore"
  >
    <ImListItem
      v-for="s in activeSessions.list"
      :key="s.id"
      :title="supportSessionTitle(s)"
      :preview="sessionAssigneePreview(s)"
      :time="imFormatTime(s.lastActiveAt)"
      :avatar-url="s.userAvatarUrl"
      :active="activeId === s.id"
      @click="emit('select', s)"
    />
  </SupportSessionList>

  <SupportSessionList
    v-else
    :has-more="closedSessions.hasMore"
    :loading-more="closedSessions.loadingMore"
    :empty="!closedSessions.loading && closedSessions.list.length === 0"
    empty-text="暂无结案记录"
    @load-more="closedSessions.loadMore"
  >
    <ImListItem
      v-for="s in closedSessions.list"
      :key="s.id"
      :title="supportSessionTitle(s)"
      :preview="sessionAssigneePreview(s)"
      :time="imFormatTime(s.closeAt || s.lastActiveAt)"
      :avatar-url="s.userAvatarUrl"
      :active="activeId === s.id"
      @click="emit('select', s)"
    />
  </SupportSessionList>
</template>

<script setup lang="ts">
import type { SupportAdminSession } from '@/api/support'
import {
  supportSessionTitle,
  type SupportSessionList as SessionList,
} from '@/composables/useSupportSessions'
import ImListItem from '@/components/im/ImListItem.vue'
import { imFormatTime, imFormatWait, imWaitTone } from '@/components/im/media'
import SupportSessionList from './SupportSessionList.vue'

defineProps<{
  tab: 'queue' | 'active' | 'closed'
  queue: SessionList
  queueItems: SupportAdminSession[]
  activeSessions: SessionList
  closedSessions: SessionList
  activeId?: number
  canAccept: boolean
}>()

const emit = defineEmits<{
  select: [session: SupportAdminSession]
  accept: [session: SupportAdminSession]
}>()

function sessionAssigneePreview(s: SupportAdminSession) {
  const assignee = s.assigneeName ? `坐席 ${s.assigneeName}` : '未分配坐席'
  const preview = s.lastContent || '暂无消息'
  return `${assignee} · ${preview}`
}
</script>
