<template>
  <el-dialog v-model="visible" :title="title" width="640px" append-to-body destroy-on-close>
    <div class="support-transcript__body im-vars">
      <ImMessageThread
        :messages="messages"
        :self-id="0"
        mine-mode="support"
        :loading="loading"
        :has-more="hasMore"
        :loading-more="loadingMore"
        @load-more="loadMore"
      />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getSupportMessages, type SupportMessageItem } from '@/api/support'
import ImMessageThread from '@/components/im/ImMessageThread.vue'
import { showApiError } from '@/utils/apiError'

const LIMIT = 50

const props = defineProps<{ conversationId?: number; title: string }>()
const visible = defineModel<boolean>({ required: true })

const messages = ref<SupportMessageItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const hasMore = ref(false)

async function load() {
  const id = props.conversationId
  if (!id) return
  loading.value = true
  messages.value = []
  hasMore.value = false
  try {
    const { data } = await getSupportMessages({ conversationId: id, limit: LIMIT })
    if (id !== props.conversationId) return
    messages.value = data.list || []
    hasMore.value = messages.value.length >= LIMIT
  } catch (e) {
    showApiError(e, '加载会话记录失败')
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  const id = props.conversationId
  const firstId = messages.value[0]?.id
  if (!id || !firstId || loadingMore.value) return
  loadingMore.value = true
  try {
    const { data } = await getSupportMessages({
      conversationId: id,
      beforeId: firstId,
      limit: LIMIT,
    })
    const older = data.list || []
    hasMore.value = older.length >= LIMIT
    const seen = new Set(messages.value.map((m) => m.id))
    messages.value = [...older.filter((m) => !seen.has(m.id)), ...messages.value]
  } catch (e) {
    showApiError(e, '加载更早消息失败')
  } finally {
    loadingMore.value = false
  }
}

watch(
  () => [visible.value, props.conversationId] as const,
  ([open]) => {
    if (open) void load()
  },
)
</script>

<style src="@/components/im/im-sidebar.css"></style>
<style src="@/components/im/im-thread.css"></style>

<style scoped>
.support-transcript__body {
  display: flex;
  flex-direction: column;
  height: 60vh;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}
</style>
