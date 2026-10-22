<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { Document } from '@element-plus/icons-vue'
import type { ImMineMode, ImThreadMessage } from './types'
import { imConfirmDownload, imFileName, imFormatTime, imInitials, imMessagePreview } from './media'
import { MessageContentType } from '@/constants/message'
import { parseMessageMediaUrl } from '@/utils/messageDisplay'

const props = withDefaults(
  defineProps<{
    messages: ImThreadMessage[]
    selfId: number
    mineMode?: ImMineMode
    loading?: boolean
    hasMore?: boolean
    loadingMore?: boolean
    showRecall?: boolean
  }>(),
  {
    mineMode: 'staff',
    loading: false,
    hasMore: false,
    loadingMore: false,
    showRecall: false,
  },
)

const emit = defineEmits<{
  loadMore: []
  recall: [msg: ImThreadMessage]
  resend: [msg: ImThreadMessage]
}>()

const box = ref<HTMLElement | null>(null)

function isSystem(m: ImThreadMessage) {
  return m.senderKind === 3
}

function isMine(m: ImThreadMessage) {
  if (isSystem(m)) return false
  if (props.mineMode === 'support') {
    return m.senderKind === 2 && m.senderId === props.selfId
  }
  return m.senderId === props.selfId
}

function canRecall(m: ImThreadMessage) {
  return props.showRecall && isMine(m) && !m.recalled && m.id > 0 && !m.sendState
}

function messageKey(m: ImThreadMessage) {
  return m.clientMsgId ? `c-${m.clientMsgId}` : `m-${m.id}`
}

function scrollBottom() {
  const el = box.value
  if (el) el.scrollTop = el.scrollHeight
}

watch(
  () => props.messages[props.messages.length - 1]?.id,
  async (lastId, prevLastId) => {
    if (lastId == null) return
    if (prevLastId !== undefined && lastId === prevLastId) return
    await nextTick()
    scrollBottom()
  },
)

// 向上加载更早消息时保持当前阅读位置。
let prependAnchor: { height: number; top: number } | null = null
const edgeIds = () =>
  [props.messages[0]?.id, props.messages[props.messages.length - 1]?.id] as const
watch(
  edgeIds,
  ([first, last], [prevFirst, prevLast]) => {
    const el = box.value
    const prepended = prevFirst !== undefined && first !== prevFirst && last === prevLast
    prependAnchor = el && prepended ? { height: el.scrollHeight, top: el.scrollTop } : null
  },
  { flush: 'pre' },
)
watch(
  edgeIds,
  () => {
    const el = box.value
    if (!el || !prependAnchor) return
    el.scrollTop = el.scrollHeight - prependAnchor.height + prependAnchor.top
    prependAnchor = null
  },
  { flush: 'post' },
)

defineExpose({ scrollBottom })
</script>

<template>
  <div ref="box" class="im-thread" v-loading="loading">
    <div v-if="hasMore" class="im-thread__load">
      <el-button link type="primary" :loading="loadingMore" @click="emit('loadMore')">
        加载更早消息
      </el-button>
    </div>

    <template v-for="m in messages" :key="messageKey(m)">
      <div v-if="isSystem(m)" class="im-msg is-system">
        <div class="im-msg__system">{{ m.content }}</div>
      </div>
      <div v-else class="im-msg" :class="{ 'is-mine': isMine(m) }">
        <el-avatar :size="34" :src="m.senderAvatarUrl || undefined">
          {{ imInitials(m.senderName) }}
        </el-avatar>
        <div class="im-msg__col">
          <div class="im-msg__meta">
            <span v-if="!isMine(m)" class="im-msg__name">{{ m.senderName || '未知' }}</span>
            <span class="im-msg__time">{{ imFormatTime(m.createdAt) }}</span>
          </div>
          <div
            class="im-bubble"
            :class="{ 'is-recalled': m.recalled, 'is-pending': m.sendState === 'sending' }"
          >
            <template v-if="m.recalled">已撤回</template>
            <template v-else-if="m.contentType === MessageContentType.Image">
              <el-image
                class="im-bubble__image"
                :src="parseMessageMediaUrl(m.content)"
                fit="cover"
                :preview-src-list="[parseMessageMediaUrl(m.content)]"
              />
            </template>
            <template v-else-if="m.contentType === MessageContentType.Video">
              <video
                class="im-bubble__video"
                :src="parseMessageMediaUrl(m.content)"
                controls
                preload="metadata"
              />
            </template>
            <template v-else-if="m.contentType === MessageContentType.File">
              <button type="button" class="im-bubble__file" @click="imConfirmDownload(m.content)">
                <span class="im-bubble__file-icon">
                  <el-icon :size="18"><Document /></el-icon>
                </span>
                <span class="im-bubble__file-name">{{ imFileName(m.content) }}</span>
              </button>
            </template>
            <template v-else>{{ imMessagePreview(m) }}</template>
          </div>
          <div v-if="m.sendState === 'sending'" class="im-msg__state">发送中…</div>
          <div v-else-if="m.sendState === 'failed'" class="im-msg__state is-failed">
            发送失败
            <el-button link type="primary" size="small" @click="emit('resend', m)">重发</el-button>
          </div>
          <div v-else-if="canRecall(m)" class="im-msg__ops">
            <el-button link size="small" @click="emit('recall', m)">撤回</el-button>
          </div>
        </div>
      </div>
    </template>

    <el-empty v-if="!loading && messages.length === 0" description="暂无消息" :image-size="72" />
  </div>
</template>
