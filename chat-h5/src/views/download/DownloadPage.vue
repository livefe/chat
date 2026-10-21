<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'

import { getAppReleaseCurrent, type AppReleasePublicItem } from '@/api/appRelease'
import { SHARE_QUERY, normalizeAppPath, parseShareResourcePath } from '@/constants/shareLinks'
import { isInAppShell } from '@/utils/appShell'
import { tryOpenApp } from '@/utils/share/openApp'
import { safeRedirectPath } from '@/utils/safeRedirect'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const items = ref<AppReleasePublicItem[]>([])
const showWechatTip = ref(false)

const inApp = computed(() => isInAppShell(route.query))
const continuePath = computed(() => {
  const raw = route.query[SHARE_QUERY.continue]
  if (typeof raw !== 'string' || !raw.trim()) return ''
  return safeRedirectPath(normalizeAppPath(raw), '')
})
const continueResource = computed(() =>
  continuePath.value ? parseShareResourcePath(continuePath.value) : null,
)

function isWeChat() {
  return /MicroMessenger/i.test(navigator.userAgent)
}

function detectPlatform(): string {
  const ua = navigator.userAgent
  if (/iPhone|iPad|iPod/i.test(ua)) return 'ios'
  if (/Android/i.test(ua)) return 'android'
  return ''
}

const displayItems = computed(() => {
  const platform = detectPlatform()
  if (!platform || items.value.length <= 1) return items.value
  const matched = items.value.filter((i) => i.platform === platform)
  const rest = items.value.filter((i) => i.platform !== platform)
  return [...matched, ...rest]
})

async function load() {
  loading.value = true
  try {
    const res = await getAppReleaseCurrent()
    items.value = res.data?.list || []
  } catch {
    showToast('加载失败')
  } finally {
    loading.value = false
  }
}

function handleDownload(item: AppReleasePublicItem) {
  if (inApp.value) {
    showToast('已在 App 内')
    return
  }
  if (!item.downloadUrl) {
    showToast('暂无安装包')
    return
  }
  if (isWeChat()) {
    showWechatTip.value = true
    return
  }
  window.location.href = item.downloadUrl
}

function handleOpenApp() {
  if (inApp.value) {
    showToast('已在 App 内')
    return
  }
  if (isWeChat()) {
    showWechatTip.value = true
    return
  }
  const resource = continueResource.value
  if (resource && resource.type !== 'download') {
    tryOpenApp(resource)
    return
  }
  tryOpenApp({ type: 'download', continuePath: continuePath.value || undefined })
}

function handleContinueWeb() {
  if (!continuePath.value) {
    router.replace('/chat')
    return
  }
  router.replace(continuePath.value)
}

onMounted(load)
</script>

<template>
  <div class="download-page">
    <div class="hero">
      <div class="logo">聚点</div>
      <div class="subtitle">发现身边的人和事</div>
      <p v-if="continuePath" class="continue-hint">安装后将回到你分享的内容</p>
    </div>

    <div class="panel">
      <template v-if="inApp">
        <van-button block type="primary" disabled>已在 App 内使用</van-button>
      </template>
      <template v-else>
        <van-button
          v-if="continuePath"
          block
          type="primary"
          class="open-btn"
          @click="handleOpenApp"
        >
          已安装？打开 App
        </van-button>
        <template v-if="loading">
          <div class="pkg pkg--state">加载中…</div>
        </template>
        <template v-else-if="displayItems.length === 0">
          <div class="pkg pkg--state">暂无安装包</div>
        </template>
        <template v-else>
          <div class="pkg-grid" :class="{ 'pkg-grid--single': displayItems.length === 1 }">
            <button
              v-for="item in displayItems"
              :key="item.platform"
              type="button"
              class="pkg"
              @click="handleDownload(item)"
            >
              <span class="pkg-name">{{ item.platformText }}</span>
              <span class="pkg-ver">{{ item.version }}</span>
              <span class="pkg-action">下载</span>
            </button>
          </div>
        </template>
        <button v-if="continuePath" type="button" class="web-continue" @click="handleContinueWeb">
          先在网页查看
        </button>
      </template>
    </div>

    <div v-if="showWechatTip" class="wechat-mask" @click="showWechatTip = false">
      <div class="wechat-tip">
        <p>请点击右上角</p>
        <p>选择「在浏览器中打开」后下载</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.download-page {
  min-height: var(--app-vh, 100dvh);
  background: linear-gradient(180deg, #f0f0f0 0%, #ffffff 55%);
  padding: 24px 24px 48px;
  position: relative;
}
.hero {
  text-align: center;
  padding: 48px 0 36px;
}
.logo {
  font-size: 36px;
  font-weight: 600;
  color: var(--brand-color-dark);
  letter-spacing: 0.5px;
}
.subtitle {
  margin-top: 12px;
  font-size: 14px;
  color: var(--text-muted);
}
.continue-hint {
  margin: 14px 0 0;
  font-size: 13px;
  color: var(--text-secondary, #666);
}
.panel {
  max-width: 320px;
  margin: 0 auto;
}
.open-btn {
  margin-bottom: 14px;
}
.pkg-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.pkg-grid--single {
  grid-template-columns: 1fr;
  max-width: 200px;
  margin: 0 auto;
}
.pkg {
  aspect-ratio: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px 14px;
  border: none;
  border-radius: 16px;
  background: var(--card-bg);
  box-shadow: var(--shadow-card);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  font: inherit;
}
.pkg:active {
  background: #f7f7f7;
}
.pkg--state {
  aspect-ratio: auto;
  min-height: 140px;
  color: var(--text-muted);
  font-size: 14px;
  cursor: default;
  box-shadow: var(--shadow-card);
  border-radius: 16px;
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
}
.pkg-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.2;
}
.pkg-ver {
  font-size: 14px;
  color: var(--text-muted);
  line-height: 1.2;
}
.pkg-action {
  margin-top: 6px;
  padding: 8px 20px;
  border-radius: 999px;
  background: var(--brand-color);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
}
.web-continue {
  display: block;
  width: 100%;
  margin-top: 18px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 14px;
  padding: 8px;
  cursor: pointer;
}
.wechat-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.72);
  z-index: 100;
  padding: 24px;
}
.wechat-tip {
  margin-top: 48px;
  text-align: right;
  color: #fff;
  font-size: 15px;
  line-height: 1.7;
}
</style>
