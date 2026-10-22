<template>
  <aside class="support-history">
    <div class="support-history__hd">
      <span>历史咨询</span>
      <span v-if="total" class="support-history__count">{{ total }}</span>
    </div>
    <div v-loading="loading" class="support-history__list">
      <button
        v-for="s in list"
        :key="s.id"
        type="button"
        class="support-history__item"
        @click="openTranscript(s)"
      >
        <div class="support-history__row">
          <span class="support-history__date">{{ s.createdAt }}</span>
          <el-tag
            size="small"
            :type="s.status === SupportSessionStatus.Closed ? 'info' : 'success'"
          >
            {{ s.statusText }}
          </el-tag>
        </div>
        <div class="support-history__meta">
          {{ s.assigneeName ? `坐席 ${s.assigneeName}` : '未接入' }}
          <template v-if="s.closeReasonText"> · {{ s.closeReasonText }}</template>
        </div>
        <div v-if="s.closeNote" class="support-history__note">备注：{{ s.closeNote }}</div>
        <div class="support-history__preview">{{ s.lastMessage || '暂无消息' }}</div>
      </button>
      <div v-if="!loading && !list.length" class="support-history__empty">
        {{ error ? '加载失败' : '没有其他咨询记录' }}
        <el-button v-if="error" link type="primary" @click="load(false)">重试</el-button>
      </div>
      <div v-if="list.length < total" class="support-history__more">
        <el-button link type="primary" :loading="loadingMore" @click="load(true)">
          加载更多
        </el-button>
      </div>
    </div>

    <SupportTranscriptDialog
      v-model="transcriptVisible"
      :conversation-id="transcriptId"
      :title="transcriptTitle"
    />
  </aside>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getSupportHistory, type SupportUserSession } from '@/api/support'
import { SupportSessionStatus } from '@/constants/support'
import SupportTranscriptDialog from './SupportTranscriptDialog.vue'

const PAGE_SIZE = 10

const props = defineProps<{ conversationId: number }>()

const list = ref<SupportUserSession[]>([])
const total = ref(0)
const loading = ref(false)
const loadingMore = ref(false)
const error = ref(false)
const transcriptVisible = ref(false)
const transcriptId = ref<number>()
const transcriptTitle = ref('')

async function load(append: boolean) {
  const convId = props.conversationId
  const page = append ? Math.floor(list.value.length / PAGE_SIZE) + 1 : 1
  if (append) loadingMore.value = true
  else loading.value = true
  error.value = false
  try {
    const { data } = await getSupportHistory({ conversationId: convId, page, pageSize: PAGE_SIZE })
    if (convId !== props.conversationId) return
    total.value = data.total
    list.value = append ? [...list.value, ...data.list] : data.list
  } catch {
    if (!append) {
      list.value = []
      total.value = 0
      error.value = true
    }
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function openTranscript(s: SupportUserSession) {
  transcriptId.value = s.id
  transcriptTitle.value = `咨询记录 · ${s.createdAt}`
  transcriptVisible.value = true
}

watch(
  () => props.conversationId,
  () => void load(false),
  { immediate: true },
)
</script>

<style scoped>
.support-history {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--im-border);
  background: #fafbfc;
  min-height: 0;
}

.support-history__hd {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 12px 14px;
  font-size: 13px;
  font-weight: 600;
  border-bottom: 1px solid var(--im-border);
}

.support-history__count {
  font-size: 12px;
  font-weight: normal;
  color: var(--im-text-secondary);
}

.support-history__list {
  flex: 1;
  overflow: auto;
  min-height: 80px;
}

.support-history__item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
  padding: 10px 14px;
  border: none;
  border-bottom: 1px solid var(--im-border);
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.support-history__item:hover {
  background: rgba(0, 0, 0, 0.035);
}

.support-history__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.support-history__date {
  font-size: 12px;
  color: var(--im-text);
}

.support-history__meta,
.support-history__note,
.support-history__preview {
  font-size: 12px;
  color: var(--im-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.support-history__empty {
  padding: 24px 14px;
  font-size: 12px;
  text-align: center;
  color: var(--im-text-secondary);
}

.support-history__more {
  display: flex;
  justify-content: center;
  padding: 8px 0;
}
</style>
