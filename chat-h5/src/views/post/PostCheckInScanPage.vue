<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'
import PostScanFinderMask from './components/PostScanFinderMask.vue'
import { postCheckIn } from '@/api/post'
import { getErrorMessage } from '@/utils/apiError'
import { buildCheckInResultPayload, encodeCheckInResultPayload } from '@/utils/post/checkInResult'

type BarcodeDetectorLike = {
  detect: (source: ImageBitmapSource) => Promise<Array<{ rawValue?: string }>>
}

const router = useRouter()

const videoRef = ref<HTMLVideoElement | null>(null)
const supported = ref(true)
const scanning = ref(false)
const busy = ref(false)
const tip = ref('将二维码放入框内')

let stream: MediaStream | null = null
let detector: BarcodeDetectorLike | null = null
let raf = 0
let lastPayload = ''
let lastAt = 0

function createDetector(): BarcodeDetectorLike | null {
  const Ctor = (
    window as Window & {
      BarcodeDetector?: new (opts?: { formats: string[] }) => BarcodeDetectorLike
    }
  ).BarcodeDetector
  if (!Ctor) return null
  try {
    return new Ctor({ formats: ['qr_code'] })
  } catch {
    return null
  }
}

async function startCamera() {
  detector = createDetector()
  if (!detector) {
    supported.value = false
    tip.value = '当前浏览器不支持扫码，请使用 App 或在报名名单页手动核验'
    return
  }
  try {
    stream = await navigator.mediaDevices.getUserMedia({
      audio: false,
      video: {
        facingMode: { ideal: 'environment' },
        width: { ideal: 1280 },
        height: { ideal: 720 },
      },
    })
    await nextTick()
    const video = videoRef.value
    if (!video) return
    video.srcObject = stream
    video.setAttribute('playsinline', 'true')
    video.muted = true
    await video.play()
    scanning.value = true
    tip.value = '将二维码放入框内'
    loop()
  } catch {
    supported.value = false
    tip.value = '无法打开摄像头，请检查权限或使用手动核验'
  }
}

function stopCamera() {
  if (raf) cancelAnimationFrame(raf)
  raf = 0
  scanning.value = false
  if (stream) {
    stream.getTracks().forEach((t) => t.stop())
    stream = null
  }
  const video = videoRef.value
  if (video) video.srcObject = null
}

function loop() {
  raf = requestAnimationFrame(async () => {
    if (!scanning.value || !detector || !videoRef.value) {
      if (scanning.value) loop()
      return
    }
    if (busy.value) {
      loop()
      return
    }
    const video = videoRef.value
    if (video.readyState < 2) {
      loop()
      return
    }
    try {
      const codes = await detector.detect(video)
      const payload = codes.find((c) => c.rawValue)?.rawValue?.trim()
      if (payload) await handlePayload(payload)
    } catch {}
    if (scanning.value) loop()
  })
}

function goResult(payload: ReturnType<typeof buildCheckInResultPayload>) {
  router.push({
    path: '/check-in-result',
    query: { payload: encodeCheckInResultPayload(payload) },
  })
}

async function handlePayload(payload: string) {
  const now = Date.now()
  if (payload === lastPayload && now - lastAt < 1800) return
  if (busy.value) return
  busy.value = true
  lastPayload = payload
  lastAt = now
  tip.value = '核验中…'
  try {
    const res = await postCheckIn({ payload })
    if (res.data.alreadyCheckedIn) {
      goResult(buildCheckInResultPayload('already', res.data))
    } else {
      goResult(buildCheckInResultPayload('success', res.data))
    }
  } catch (e: unknown) {
    goResult(buildCheckInResultPayload('fail', undefined, getErrorMessage(e, '核验失败')))
  } finally {
    busy.value = false
    tip.value = '将二维码放入框内'
  }
}

onMounted(() => {
  void startCamera()
})

onBeforeUnmount(() => {
  stopCamera()
})
</script>

<template>
  <div class="scan-shell">
    <SubPageLayout full-height>
      <template #title>扫一扫</template>
      <div class="scan-page">
        <div class="scan-viewport">
          <video ref="videoRef" class="scan-video" playsinline muted autoplay />
          <div class="scan-top-scrim" aria-hidden="true" />
          <PostScanFinderMask />
          <div class="scan-tip-overlay">
            <div class="scan-tip">{{ tip }}</div>
            <p v-if="!supported" class="scan-fallback">
              可返回报名名单，对未核验用户使用「手动核验」。
            </p>
          </div>
        </div>
      </div>
    </SubPageLayout>
  </div>
</template>

<style scoped>
.scan-shell {
  height: var(--app-vh, 100dvh);
}

.scan-shell :deep(.sub-page) {
  position: relative;
  background: #000;
}

.scan-shell :deep(.sub-page__body) {
  position: relative;
}

.scan-shell :deep(.van-nav-bar),
.scan-shell :deep(.van-nav-bar__content) {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 3;
  background: transparent !important;
}

.scan-shell :deep(.van-nav-bar .van-icon),
.scan-shell :deep(.van-nav-bar__title),
.scan-shell :deep(.van-nav-bar__text) {
  color: #fff;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.45);
}

.scan-page {
  position: absolute;
  inset: 0;
  color: #fff;
}

.scan-viewport {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: #000;
}

.scan-video {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.scan-top-scrim {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 2;
  height: calc(env(safe-area-inset-top) + 72px);
  pointer-events: none;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.32) 0%, rgba(0, 0, 0, 0) 100%);
}

.scan-tip-overlay {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 2;
  padding: 28px 20px calc(18px + env(safe-area-inset-bottom));
  text-align: center;
  pointer-events: none;
  background: linear-gradient(
    180deg,
    rgba(0, 0, 0, 0) 0%,
    rgba(0, 0, 0, 0.28) 36%,
    rgba(0, 0, 0, 0.5) 100%
  );
}

.scan-tip {
  font-size: 15px;
}

.scan-fallback {
  margin: 10px 0 0;
  text-align: center;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.65);
  line-height: 1.5;
}
</style>
