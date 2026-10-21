<script setup lang="ts">
import { computed, ref } from 'vue'
import type { AdminMessageItem } from '@/api/message'
import { MessageContentType } from '@/constants/message'
import { adminReplyLabel, parseMessageMediaUrl } from '@/utils/messageDisplay'

const props = defineProps<{ item: AdminMessageItem }>()

const mediaUrl = computed(() => {
  const type = props.item.replyToContentType
  if (type !== MessageContentType.Image && type !== MessageContentType.Video) return ''
  return parseMessageMediaUrl(props.item.replyToContent)
})
const showImage = computed(
  () => props.item.replyToContentType === MessageContentType.Image && !!mediaUrl.value,
)
const showVideo = computed(
  () => props.item.replyToContentType === MessageContentType.Video && !!mediaUrl.value,
)
const videoPreview = ref(false)
</script>

<template>
  <div v-if="item.replyToId" class="reply-quote">
    <span>引用 {{ item.replyToFromName || '消息' }}：</span>
    <el-image
      v-if="showImage"
      :src="mediaUrl"
      :preview-src-list="[mediaUrl]"
      fit="cover"
      class="reply-thumb"
    />
    <button v-else-if="showVideo" type="button" class="reply-video" @click="videoPreview = true">
      <video :src="mediaUrl" class="reply-thumb" muted preload="metadata" />
      <span class="reply-video__play" aria-hidden="true" />
    </button>
    <span v-else>{{ adminReplyLabel(item) }}</span>
    <el-tag v-if="item.replyToRecalled" type="warning" size="small">已撤回</el-tag>
    <el-tag v-if="item.replyToDeleted" type="danger" size="small">已删除</el-tag>
    <el-dialog v-model="videoPreview" title="视频" width="640px" append-to-body destroy-on-close>
      <video v-if="mediaUrl" :src="mediaUrl" class="reply-video-player" controls autoplay />
    </el-dialog>
  </div>
</template>

<style scoped>
.reply-quote {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  padding: 4px 8px;
  border-left: 3px solid var(--el-color-primary-light-5);
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.4;
  word-break: break-word;
}
.reply-thumb {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  vertical-align: middle;
  object-fit: cover;
  background: var(--el-fill-color-light);
}
.reply-video {
  position: relative;
  display: inline-flex;
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
}
.reply-video .reply-thumb {
  display: block;
  pointer-events: none;
}
.reply-video__play {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 5px 0 5px 8px;
  border-color: transparent transparent transparent #fff;
  transform: translate(-40%, -50%);
  pointer-events: none;
}
.reply-video-player {
  display: block;
  width: 100%;
  max-height: 70vh;
  background: #000;
}
</style>
