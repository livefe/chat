<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

const props = defineProps<{
  downloadHref: string

  dismissKey?: string
}>()

const emit = defineEmits<{ openApp: [] }>()

const STORAGE_PREFIX = 'share_open_app_bar_dismissed:'
const visible = ref(false)
const showWechatTip = ref(false)

const isWeChat = computed(() => /MicroMessenger/i.test(navigator.userAgent))
const storageKey = computed(() => STORAGE_PREFIX + (props.dismissKey || 'default'))

onMounted(() => {
  try {
    if (sessionStorage.getItem(storageKey.value) === '1') return
  } catch {}
  visible.value = true
})

function dismiss() {
  visible.value = false
  try {
    sessionStorage.setItem(storageKey.value, '1')
  } catch {}
}

function onOpen() {
  if (isWeChat.value) {
    showWechatTip.value = true
    return
  }
  emit('openApp')
}

function onDownload() {
  if (isWeChat.value) {
    showWechatTip.value = true
    return
  }
  window.location.href = props.downloadHref
}
</script>

<template>
  <Teleport to="body">
    <Transition name="share-open">
      <div v-if="visible" class="share-open" role="region" aria-label="打开聚点 App">
        <div class="share-open__card">
          <button type="button" class="share-open__close" aria-label="关闭" @click="dismiss">
            <van-icon name="cross" />
          </button>
          <div class="share-open__mark" aria-hidden="true">聚</div>
          <div class="share-open__body">
            <div class="share-open__title">聚点 App</div>
            <div class="share-open__desc">打开 App 获得完整体验</div>
          </div>
          <div class="share-open__actions">
            <button type="button" class="share-open__link" @click="onDownload">下载</button>
            <button type="button" class="share-open__cta" @click="onOpen">打开</button>
          </div>
        </div>
      </div>
    </Transition>

    <div v-if="showWechatTip" class="share-wechat-mask" @click="showWechatTip = false">
      <div class="share-wechat-tip">
        <p>请点击右上角</p>
        <p>选择「在浏览器中打开」后继续</p>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.share-open {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: calc(12px + env(safe-area-inset-bottom, 0px));
  z-index: 40;
  pointer-events: none;
}

.share-open__card {
  pointer-events: auto;
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 12px 12px 14px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow:
    0 8px 28px rgba(0, 0, 0, 0.1),
    0 2px 8px rgba(0, 0, 0, 0.04);
  backdrop-filter: blur(10px);
}

.share-open__close {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.share-open__mark {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--brand-color);
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.02em;
}

.share-open__body {
  flex: 1;
  min-width: 0;
  padding-right: 8px;
}

.share-open__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
}

.share-open__desc {
  margin-top: 2px;
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.3;
}

.share-open__actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  padding-right: 18px;
}

.share-open__link {
  border: none;
  background: transparent;
  color: var(--text-secondary, #666);
  font: inherit;
  font-size: 13px;
  padding: 8px 6px;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.share-open__cta {
  border: none;
  border-radius: 999px;
  background: var(--brand-color);
  color: #fff;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 14px;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

.share-open-enter-active,
.share-open-leave-active {
  transition:
    opacity 0.22s ease,
    transform 0.22s ease;
}
.share-open-enter-from,
.share-open-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

.share-wechat-mask {
  position: fixed;
  inset: 0;
  z-index: 100;
  background: rgba(0, 0, 0, 0.72);
  padding: 24px;
}
.share-wechat-tip {
  margin-top: 48px;
  text-align: right;
  color: #fff;
  font-size: 15px;
  line-height: 1.7;
}
</style>
