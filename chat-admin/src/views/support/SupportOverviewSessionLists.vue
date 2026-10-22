<template>
  <SupportSessionList
    v-if="tab === 'queue'"
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
      :active="activeId === s.id"
      @click="emit('select', s)"
    />
  </SupportSessionList>

  <SupportSessionList
    v-else-if="tab === 'active'"
    :has-more="activeSessions.hasMore"
    :loading-more="activeSessions.loadingMore"
    :empty="activeSessions.loaded && activeSessions.list.length === 0"
    empty-text="暂无进行中会话"
    @load-more="activeSessions.loadMore"
  >
    <ImListItem
      v-for="s in activeSessions.list"
      :key="s.id"
      :title="supportSessionTitle(s)"
      :preview="sessionPreview(s)"
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
    :empty="closedSessions.loaded && closedSessions.list.length === 0"
    empty-text="暂无结案记录"
    @load-more="closedSessions.loadMore"
  >
    <ImListItem
      v-for="s in closedSessions.list"
      :key="s.id"
      :title="supportSessionTitle(s)"
      :preview="sessionPreview(s)"
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
  supportWaitSec,
  type SupportSessionList as SessionList,
} from '@/composables/useSupportSessions'
import ImListItem from '@/components/im/ImListItem.vue'
import { imFormatTime, imFormatWait, imWaitTone } from '@/components/im/media'
import SupportSessionList from './SupportSessionList.vue'

defineProps<{
  tab: 'queue' | 'active' | 'closed'
  queue: SessionList
  activeSessions: SessionList
  closedSessions: SessionList
  activeId?: number
  now: number
}>()

const emit = defineEmits<{
  select: [session: SupportAdminSession]
}>()

function sessionPreview(s: SupportAdminSession) {
  const parts = [s.assigneeName ? `坐席 ${s.assigneeName}` : '未分配坐席']
  if (s.closeReasonText) parts.push(s.closeReasonText)
  parts.push(s.lastContent || '暂无消息')
  return parts.join(' · ')
}
</script>
