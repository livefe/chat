<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

import type { MessageReplyQuote } from '@/api/message'
import ChatVideoPreview from './ChatVideoPreview.vue'
import FileClipIcon from './FileClipIcon.vue'
import MessageComposerReplyBar from './composer/MessageComposerReplyBar.vue'
import { useChatMediaUpload, type ChatMediaPayload } from '@/composables/useChatMediaUpload'
import { useVoiceRecorder } from '@/composables/useVoiceRecorder'
import { MessageContentTypes } from '@/constants/chat'

const props = withDefaults(
  defineProps<{
    canCompose?: boolean
    enableVoice?: boolean
    enableLocation?: boolean
    enableFile?: boolean
    placeholder?: string
  }>(),
  {
    canCompose: true,
    enableVoice: true,
    enableLocation: true,
    enableFile: true,
    placeholder: '说点什么…',
  },
)

const moreOpen = defineModel<boolean>('moreOpen', { default: false })
const replyTo = defineModel<MessageReplyQuote | null>('replyTo', { default: null })

const emit = defineEmits<{
  send: [content: string, contentType?: number, reply?: MessageReplyQuote | null]
  pickLocation: []
  stick: [force: boolean]
}>()

const COMPOSER_FIELD_MIN = 40

const input = ref('')
const voiceMode = ref(false)
const fieldHeight = ref(COMPOSER_FIELD_MIN)
const mediaInputRef = ref<HTMLInputElement | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const composerInputRef = ref<HTMLTextAreaElement | null>(null)
const replyVideoPreview = ref('')

const {
  uploadingMedia: uploadingImage,
  uploadingVoice,
  uploadMedia,
  uploadFile,
  uploadVoice,
} = useChatMediaUpload()

const {
  recording,
  recordCanceling,
  startRecording,
  cancelRecording,
  onHoldPointerMove,
  onHoldPointerUp,
  onHoldPointerCancel,
} = useVoiceRecorder({
  canStart: () =>
    props.canCompose && !uploadingImage.value && !uploadingVoice.value && voiceMode.value,
  onStart: () => {
    moreOpen.value = false
    emit('stick', true)
  },
  onRecorded: async (blob, mime, durationMs) => {
    sendPayload(await uploadVoice(blob, mime, durationMs))
  },
})

function takeReply(): MessageReplyQuote | null {
  const q = replyTo.value
  replyTo.value = null
  return q
}

function sendPayload(payload: ChatMediaPayload | null) {
  if (!payload) return
  emit('stick', true)
  emit('send', payload.content, payload.contentType, takeReply())
}

function resizeComposerInput() {
  if (voiceMode.value || !input.value) {
    fieldHeight.value = COMPOSER_FIELD_MIN
    const el = composerInputRef.value
    if (el) el.style.height = ''
    return
  }
  const el = composerInputRef.value
  if (!el) {
    fieldHeight.value = COMPOSER_FIELD_MIN
    return
  }
  const maxPx = Math.min(window.innerHeight * 0.3, 280)

  el.style.height = '0px'
  const next = Math.min(Math.max(el.scrollHeight, COMPOSER_FIELD_MIN), maxPx)
  fieldHeight.value = next
  el.style.height = '100%'
}

function onSend() {
  const text = input.value.trim()
  if (!text || !props.canCompose) return
  moreOpen.value = false
  emit('stick', true)
  emit('send', text, MessageContentTypes.Text, takeReply())
  input.value = ''
  void nextTick(() => {
    resizeComposerInput()
    emit('stick', true)
  })
}

watch(input, (value) => {
  if (value.trim() && moreOpen.value) moreOpen.value = false
  void nextTick(() => {
    resizeComposerInput()
    emit('stick', false)
  })
})

watch(
  () => props.canCompose,
  (ok) => {
    if (!ok) {
      moreOpen.value = false
      cancelRecording(true)
    }
  },
)

watch(voiceMode, (on) => {
  if (on) {
    moreOpen.value = false
    composerInputRef.value?.blur()
    fieldHeight.value = COMPOSER_FIELD_MIN
    const el = composerInputRef.value
    if (el) el.style.height = ''
  } else {
    cancelRecording(true)
    void nextTick(() => resizeComposerInput())
  }
})

function dismiss() {
  composerInputRef.value?.blur()
  moreOpen.value = false
}

function onComposerFocus() {
  moreOpen.value = false
  emit('stick', true)
}

function toggleMore() {
  if (uploadingImage.value || uploadingVoice.value || !props.canCompose) return
  moreOpen.value = !moreOpen.value
  if (moreOpen.value) emit('stick', true)
}

function toggleVoiceMode() {
  if (!props.enableVoice) return
  if (uploadingImage.value || uploadingVoice.value || recording.value || !props.canCompose) return
  voiceMode.value = !voiceMode.value
  moreOpen.value = false
}

function pickMedia() {
  if (uploadingImage.value || !props.canCompose) return
  moreOpen.value = false
  mediaInputRef.value?.click()
}

function pickFile() {
  if (uploadingImage.value || !props.canCompose || !props.enableFile) return
  moreOpen.value = false
  fileInputRef.value?.click()
}

function onPickLocation() {
  if (uploadingImage.value || !props.canCompose || !props.enableLocation) return
  moreOpen.value = false
  emit('pickLocation')
}

function takeSelectedFile(ev: Event): File | undefined {
  const inputEl = ev.target as HTMLInputElement
  const file = inputEl.files?.[0]
  inputEl.value = ''
  return file
}

async function onMediaSelected(ev: Event) {
  const file = takeSelectedFile(ev)
  if (!file) return
  sendPayload(await uploadMedia(file))
}

async function onFileSelected(ev: Event) {
  const file = takeSelectedFile(ev)
  if (!file) return
  sendPayload(await uploadFile(file))
}

function focus() {
  if (voiceMode.value) return
  void nextTick(() => composerInputRef.value?.focus())
}

defineExpose({ dismiss, focus })
</script>

<template>
  <div class="chat-composer-wrap">
    <input
      ref="mediaInputRef"
      type="file"
      accept="image/*,video/*"
      class="chat-composer__file"
      @change="onMediaSelected"
    />
    <input
      v-if="enableFile"
      ref="fileInputRef"
      type="file"
      accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.csv,.zip,.rar,.7z"
      class="chat-composer__file"
      @change="onFileSelected"
    />
    <MessageComposerReplyBar
      v-if="replyTo"
      :reply="replyTo"
      @close="replyTo = null"
      @preview-video="replyVideoPreview = $event"
    />
    <div class="chat-composer">
      <button
        v-if="enableVoice"
        type="button"
        class="chat-composer__mode"
        :disabled="uploadingImage || uploadingVoice || recording || !canCompose"
        :aria-label="voiceMode ? '切换到键盘' : '切换到语音'"
        @click="toggleVoiceMode"
      >
        <svg
          v-if="voiceMode"
          class="chat-composer__mode-icon"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            fill="currentColor"
            d="M4 5h16a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2Zm0 2v10h16V7H4Zm2 1.5h1.8v1.8H6V8.5Zm3.1 0h1.8v1.8H9.1V8.5Zm3.1 0h1.8v1.8h-1.8V8.5Zm3.1 0H18v1.8h-1.8V8.5ZM6 11.1h1.8v1.8H6v-1.8Zm3.1 0h1.8v1.8H9.1v-1.8Zm3.1 0h1.8v1.8h-1.8v-1.8Zm3.1 0H18v1.8h-1.8v-1.8ZM8.2 14.2H15.8v1.8H8.2v-1.8Z"
          />
        </svg>
        <svg
          v-else
          class="chat-composer__mode-icon"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <rect
            x="8"
            y="2.5"
            width="8"
            height="12"
            rx="4"
            stroke="currentColor"
            stroke-width="1.5"
          />
          <path
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            d="M6.5 12.5a5.5 5.5 0 0 0 11 0M12 18v3.25M9.5 21.25h5"
          />
        </svg>
      </button>
      <div class="chat-composer__main" :style="{ height: `${fieldHeight}px` }">
        <textarea
          v-show="!voiceMode"
          ref="composerInputRef"
          v-model="input"
          class="chat-composer__input"
          rows="1"
          :placeholder="placeholder"
          @focus="onComposerFocus"
          @input="resizeComposerInput"
        />
        <button
          v-show="voiceMode"
          type="button"
          class="chat-composer__hold"
          :class="{
            'chat-composer__hold--recording': recording && !recordCanceling,
            'chat-composer__hold--cancel': recording && recordCanceling,
          }"
          :disabled="uploadingVoice || !canCompose"
          @pointerdown.prevent="startRecording"
          @pointermove="onHoldPointerMove"
          @pointerup="onHoldPointerUp"
          @pointercancel="onHoldPointerCancel"
        >
          <template v-if="uploadingVoice">发送中…</template>
          <template v-else-if="recording && recordCanceling">松开手指，取消发送</template>
          <template v-else-if="recording">松开 结束</template>
          <template v-else>按住 说话</template>
        </button>
      </div>
      <button
        v-if="!voiceMode && !input.trim()"
        type="button"
        class="chat-composer__more"
        :disabled="uploadingImage || uploadingVoice || !canCompose"
        aria-label="更多"
        @click="toggleMore"
      >
        <span class="chat-composer__more-ring">
          <van-icon :name="moreOpen ? 'cross' : 'plus'" size="15" />
        </span>
      </button>
      <button
        v-else-if="!voiceMode"
        class="chat-composer__send"
        :disabled="!canCompose"
        @click="onSend"
      >
        发送
      </button>
      <button
        v-else
        type="button"
        class="chat-composer__more"
        :disabled="uploadingImage || uploadingVoice || recording || !canCompose"
        aria-label="更多"
        @click="toggleMore"
      >
        <span class="chat-composer__more-ring">
          <van-icon :name="moreOpen ? 'cross' : 'plus'" size="15" />
        </span>
      </button>
    </div>
    <div v-if="moreOpen" class="chat-composer__panel">
      <button
        type="button"
        class="chat-composer__action"
        :disabled="uploadingImage || !canCompose"
        @click="pickMedia"
      >
        <span class="chat-composer__action-icon">
          <van-icon name="photo-o" size="26" />
        </span>
        <span class="chat-composer__action-label">相册</span>
      </button>
      <button
        v-if="enableLocation"
        type="button"
        class="chat-composer__action"
        :disabled="uploadingImage || !canCompose"
        @click="onPickLocation"
      >
        <span class="chat-composer__action-icon">
          <van-icon name="location-o" size="26" />
        </span>
        <span class="chat-composer__action-label">位置</span>
      </button>
      <button
        v-if="enableFile"
        type="button"
        class="chat-composer__action"
        :disabled="uploadingImage || !canCompose"
        @click="pickFile"
      >
        <span class="chat-composer__action-icon">
          <FileClipIcon :size="26" />
        </span>
        <span class="chat-composer__action-label">文件</span>
      </button>
    </div>
    <ChatVideoPreview
      v-if="replyVideoPreview"
      :src="replyVideoPreview"
      @close="replyVideoPreview = ''"
    />
  </div>
</template>
