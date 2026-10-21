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

defineExpose({ scrollBottom })
</script>

<template>
  <div ref="box" class="im-thread" v-loading="loading">
    <div v-if="hasMore" class="im-thread__load">
      <el-button link type="primary" :loading="loadingMore" @click="emit('loadMore')">
        加载更早消息
      </el-button>
    </div>

    <template v-for="m in messages" :key="m.id">
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
          <div class="im-bubble" :class="{ 'is-recalled': m.recalled }">
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
          <div v-if="showRecall && isMine(m) && !m.recalled" class="im-msg__ops">
            <el-button link size="small" @click="emit('recall', m)">撤回</el-button>
          </div>
        </div>
      </div>
    </template>

    <el-empty v-if="!loading && messages.length === 0" description="暂无消息" :image-size="72" />
  </div>
</template>
