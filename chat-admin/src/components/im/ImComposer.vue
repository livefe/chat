<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ChatLineSquare, Document, Picture, VideoCamera } from '@element-plus/icons-vue'
import { MessageContentType } from '@/constants/message'
import type { ImQuickReply } from './types'

const props = withDefaults(
  defineProps<{
    modelValue: string
    sending?: boolean
    uploading?: boolean
    allowUpload?: boolean
    placeholder?: string
    quickReplies?: ImQuickReply[]
  }>(),
  {
    sending: false,
    uploading: false,
    allowUpload: true,
    placeholder: '输入消息，Enter 发送，Shift+Enter 换行',
    quickReplies: () => [],
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

const hasQuickReplies = computed(() => props.quickReplies.length > 0)
const pickerOpen = ref(false)
const pickerIndex = ref(0)
const slashMode = computed(() => props.modelValue.startsWith('/'))
const pickerItems = computed(() => {
  const kw = slashMode.value ? props.modelValue.slice(1).trim().toLowerCase() : ''
  if (!kw) return props.quickReplies
  return props.quickReplies.filter(
    (q) => q.title.toLowerCase().includes(kw) || q.content.toLowerCase().includes(kw),
  )
})

watch(
  () => props.modelValue,
  (v, prev) => {
    if (!hasQuickReplies.value) return
    if (v.startsWith('/') && !prev.startsWith('/')) openPicker()
    else if (!v.startsWith('/') && prev.startsWith('/')) pickerOpen.value = false
  },
)

watch(pickerItems, () => {
  pickerIndex.value = 0
})

function openPicker() {
  pickerIndex.value = 0
  pickerOpen.value = true
}

function togglePicker() {
  if (pickerOpen.value) pickerOpen.value = false
  else openPicker()
}

function pickQuickReply(q: ImQuickReply) {
  const draft = props.modelValue
  emit('update:modelValue', slashMode.value || !draft.trim() ? q.content : `${draft}${q.content}`)
  pickerOpen.value = false
}

function onPickerKey(e: KeyboardEvent) {
  const items = pickerItems.value
  if (e.key === 'Escape') {
    pickerOpen.value = false
    return true
  }
  if (!items.length) return false
  if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    const step = e.key === 'ArrowDown' ? 1 : -1
    pickerIndex.value = (pickerIndex.value + step + items.length) % items.length
    return true
  }
  if (e.key === 'Enter' && !e.shiftKey) {
    const item = items[pickerIndex.value]
    if (item) pickQuickReply(item)
    return true
  }
  return false
}

function onKeydown(e: Event | KeyboardEvent) {
  if (!(e instanceof KeyboardEvent)) return
  if (pickerOpen.value && onPickerKey(e)) {
    e.preventDefault()
    return
  }
  if (e.key !== 'Enter' || e.shiftKey || e.isComposing) return
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
    <div v-if="pickerOpen" class="im-quick">
      <div class="im-quick__head">
        <span>快捷回复</span>
        <span class="im-quick__hint">↑↓ 选择，Enter 插入，Esc 关闭</span>
      </div>
      <div class="im-quick__list">
        <button
          v-for="(q, i) in pickerItems"
          :key="`${q.title}-${i}`"
          type="button"
          class="im-quick__item"
          :class="{ 'is-active': i === pickerIndex }"
          @mouseenter="pickerIndex = i"
          @mousedown.prevent="pickQuickReply(q)"
        >
          <span class="im-quick__title">{{ q.title }}</span>
          <span class="im-quick__content">{{ q.content }}</span>
        </button>
        <div v-if="!pickerItems.length" class="im-quick__empty">没有匹配的快捷回复</div>
      </div>
    </div>
    <div v-if="allowUpload || hasQuickReplies" class="im-composer__tools">
      <template v-if="allowUpload">
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
      </template>
      <button
        v-if="hasQuickReplies"
        type="button"
        class="im-composer__tool im-composer__tool--text"
        title="快捷回复（输入 / 也可打开）"
        @click="togglePicker"
      >
        <el-icon><ChatLineSquare /></el-icon>
        <span>快捷回复</span>
      </button>
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
        @blur="pickerOpen = false"
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
