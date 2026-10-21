<script setup lang="ts">
import { computed, ref } from 'vue'
import { Document, Picture, VideoCamera } from '@element-plus/icons-vue'
import { MessageContentType } from '@/constants/message'

const props = withDefaults(
  defineProps<{
    modelValue: string
    sending?: boolean
    uploading?: boolean
    allowUpload?: boolean
    placeholder?: string
  }>(),
  {
    sending: false,
    uploading: false,
    allowUpload: true,
    placeholder: '输入消息，Enter 发送，Shift+Enter 换行',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  send: []
  upload: [contentType: number, file: File]
}>()

const busy = computed(() => props.sending || props.uploading)
const imageInput = ref<HTMLInputElement | null>(null)
const videoInput = ref<HTMLInputElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

function onKeydown(e: Event | KeyboardEvent) {
  if (!(e instanceof KeyboardEvent) || e.key !== 'Enter' || e.shiftKey) return
  e.preventDefault()
  if (!busy.value) emit('send')
}

function pick(kind: 'image' | 'video' | 'file') {
  if (busy.value) return
  if (kind === 'image') imageInput.value?.click()
  else if (kind === 'video') videoInput.value?.click()
  else fileInput.value?.click()
}

function onFileChange(contentType: number, e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  emit('upload', contentType, file)
}
</script>

<template>
  <div class="im-composer">
    <div v-if="allowUpload" class="im-composer__tools">
      <button
        type="button"
        class="im-composer__tool"
        title="图片"
        :disabled="busy"
        @click="pick('image')"
      >
        <el-icon><Picture /></el-icon>
      </button>
      <button
        type="button"
        class="im-composer__tool"
        title="视频"
        :disabled="busy"
        @click="pick('video')"
      >
        <el-icon><VideoCamera /></el-icon>
      </button>
      <button
        type="button"
        class="im-composer__tool"
        title="文件"
        :disabled="busy"
        @click="pick('file')"
      >
        <el-icon><Document /></el-icon>
      </button>
      <input
        ref="imageInput"
        type="file"
        accept="image/*"
        hidden
        @change="onFileChange(MessageContentType.Image, $event)"
      />
      <input
        ref="videoInput"
        type="file"
        accept="video/*"
        hidden
        @change="onFileChange(MessageContentType.Video, $event)"
      />
      <input
        ref="fileInput"
        type="file"
        accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.csv,.zip,.rar,.7z"
        hidden
        @change="onFileChange(MessageContentType.File, $event)"
      />
    </div>
    <div class="im-composer__row">
      <el-input
        :model-value="modelValue"
        type="textarea"
        :rows="3"
        :placeholder="placeholder"
        :disabled="sending && !uploading"
        @update:model-value="emit('update:modelValue', $event)"
        @keydown="onKeydown"
      />
      <el-button
        class="im-composer__send"
        type="primary"
        :loading="sending"
        :disabled="uploading || !modelValue.trim()"
        @click="emit('send')"
      >
        发送
      </el-button>
    </div>
  </div>
</template>
