<script setup lang="ts">
import { computed, ref } from 'vue'

import UserAvatar from '@/components/UserAvatar.vue'
import type { ConversationItem } from '@/api/conversation'
import { peerDisplayName } from '@/utils/user'
import ForwardTargetItem from './ForwardTargetItem.vue'
import ForwardTargetPane from './ForwardTargetPane.vue'

const props = defineProps<{
  chats: ConversationItem[]
  finished: boolean
  sending: boolean
}>()

const loading = defineModel<boolean>('loading', { required: true })

const emit = defineEmits<{
  load: []
  select: [chat: ConversationItem]
}>()

const query = ref('')

const filteredChats = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.chats
  return props.chats.filter((c) => peerDisplayName(c).toLowerCase().includes(q))
})
</script>

<template>
  <ForwardTargetPane v-model:query="query" placeholder="搜索会话" :disabled="sending">
    <van-loading v-if="loading && chats.length === 0" size="24px" />
    <van-empty v-else-if="chats.length === 0" description="暂无消息" />
    <van-empty v-else-if="filteredChats.length === 0" description="没有匹配的会话" />
    <van-list
      v-else
      v-model:loading="loading"
      class="forward-list"
      :finished="finished"
      :immediate-check="false"
      finished-text=""
      @load="emit('load')"
    >
      <ForwardTargetItem
        v-for="c in filteredChats"
        :key="c.peerId"
        :name="peerDisplayName(c)"
        :disabled="sending"
        @select="emit('select', c)"
      >
        <template #icon>
          <UserAvatar :src="c.peerAvatar" :size="40" />
        </template>
      </ForwardTargetItem>
    </van-list>
  </ForwardTargetPane>
</template>
