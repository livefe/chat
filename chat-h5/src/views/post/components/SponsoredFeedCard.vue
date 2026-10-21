<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { reportAdEvents } from '@/api/ad'
import type { SponsoredItem } from '@/api/post'
import { useLocationStore } from '@/stores/location'
import { resolveMediaUrl } from '@/utils/media'

const WATERFALL_RATIO_MIN = 0.66
const WATERFALL_RATIO_MAX = 3 / 4
const WATERFALL_RATIO_DEFAULT = 3 / 4
const IMPRESSION_RATIO = 0.5
const IMPRESSION_DWELL_MS = 1000

const props = defineProps<{
  item: SponsoredItem
}>()

const router = useRouter()
const gps = useLocationStore().gps

const rootEl = ref<HTMLElement | null>(null)
const coverUrl = computed(() => resolveMediaUrl(props.item.imageUrl))
const advertiserLabel = computed(() => (props.item.advertiserName || '').trim() || '广告主')

const adBadge = computed(() => {
  const badge = (props.item.badge || '').trim()
  return badge || '广告'
})

const coverRatio = ref(WATERFALL_RATIO_DEFAULT)
const coverReady = ref(false)

let impressionReported = false
let dwellTimer: ReturnType<typeof setTimeout> | null = null
let observer: IntersectionObserver | null = null

watch(coverUrl, () => {
  coverRatio.value = WATERFALL_RATIO_DEFAULT
  coverReady.value = false
})

function clampWaterfallRatio(w: number, h: number) {
  return Math.min(WATERFALL_RATIO_MAX, Math.max(WATERFALL_RATIO_MIN, w / h))
}

function onCoverLoad(e: Event) {
  const img = e.target as HTMLImageElement
  const w = img.naturalWidth
  const h = img.naturalHeight
  if (w > 0 && h > 0) {
    coverRatio.value = clampWaterfallRatio(w, h)
  }
  coverReady.value = true
}

function clearDwellTimer() {
  if (dwellTimer != null) {
    clearTimeout(dwellTimer)
    dwellTimer = null
  }
}

function geoFields() {
  if (!gps.valid) return {}
  return {
    latitude: gps.lat,
    longitude: gps.lng,
  }
}

function eventPayload(eventType: 'impression' | 'click') {
  return {
    requestId: props.item.requestId,
    campaignId: props.item.campaignId,
    eventType,
    impressionToken: props.item.impressionToken,
    slotIndex: props.item.slotIndex,
    ...geoFields(),
  }
}

async function reportImpressionOnce() {
  if (impressionReported) return
  impressionReported = true
  clearDwellTimer()
  try {
    await reportAdEvents([eventPayload('impression')])
  } catch {}
}

function onIntersect(entries: IntersectionObserverEntry[]) {
  const entry = entries[0]
  if (!entry) return
  if (entry.isIntersecting && entry.intersectionRatio >= IMPRESSION_RATIO) {
    if (impressionReported || dwellTimer != null) return
    dwellTimer = setTimeout(() => {
      dwellTimer = null
      void reportImpressionOnce()
    }, IMPRESSION_DWELL_MS)
  } else {
    clearDwellTimer()
  }
}

function openLanding() {
  const type = (props.item.landingType || '').trim()
  const value = (props.item.landingValue || '').trim()
  if (!value) return

  if (type === 'url') {
    window.open(value, '_blank', 'noopener,noreferrer')
    return
  }
  if (type === 'post') {
    router.push(`/post/${value}`)
    return
  }
  if (type === 'room') {
    router.push(`/room/${value}`)
  }
}

async function onCardClick() {
  try {
    await reportAdEvents([eventPayload('click')])
  } catch {}
  openLanding()
}

onMounted(() => {
  const el = rootEl.value
  if (!el || typeof IntersectionObserver === 'undefined') {
    void reportImpressionOnce()
    return
  }
  observer = new IntersectionObserver(onIntersect, {
    threshold: [0, IMPRESSION_RATIO, 1],
  })
  observer.observe(el)
})

onBeforeUnmount(() => {
  clearDwellTimer()
  observer?.disconnect()
  observer = null
})
</script>

<template>
  <article ref="rootEl" class="sponsored-card" @click="onCardClick">
    <div v-if="coverUrl" class="sponsored-card__cover" :style="{ aspectRatio: String(coverRatio) }">
      <img
        :src="coverUrl"
        alt=""
        class="sponsored-card__cover-img"
        :class="{ 'sponsored-card__cover-img--ready': coverReady }"
        loading="lazy"
        decoding="async"
        @load="onCoverLoad"
      />
    </div>
    <div v-else class="sponsored-card__text-cover">
      <p class="sponsored-card__text-cover-body">{{ item.title }}</p>
    </div>
    <div class="sponsored-card__caption">
      <h3 class="sponsored-card__title">{{ item.title }}</h3>
      <div class="sponsored-card__meta-row">
        <span class="sponsored-card__advertiser">{{ advertiserLabel }}</span>
        <span class="sponsored-card__ad-badge" aria-label="广告">{{ adBadge }}</span>
      </div>
    </div>
  </article>
</template>

<style scoped>
.sponsored-card {
  display: block;
  position: relative;
  padding: 0;
  border-radius: 8px;
  overflow: hidden;
  background: var(--card-bg, #fff);
  cursor: pointer;
}

.sponsored-card:active {
  opacity: 0.96;
}

.sponsored-card__cover {
  position: relative;
  overflow: hidden;
  background: linear-gradient(180deg, #f0f0f0 0%, #e5e5e5 100%);
}

.sponsored-card__cover-img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  opacity: 0;
  transition: opacity 0.22s ease;
}

.sponsored-card__cover-img--ready {
  opacity: 1;
}

.sponsored-card__text-cover {
  position: relative;
  box-sizing: border-box;
  aspect-ratio: 3 / 4;
  padding: 20px 18px;
  background: linear-gradient(165deg, #f0f0f0 0%, #e8e8e8 52%, #d8d8d8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.sponsored-card__text-cover-body {
  margin: 0;
  width: 100%;
  font-size: 16px;
  line-height: 1.58;
  font-weight: 500;
  letter-spacing: 0.01em;
  color: #2a2a2a;
  text-align: center;
  font-family: 'PingFang SC', 'Hiragino Sans GB', 'Noto Sans SC', 'Microsoft YaHei', sans-serif;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 6;
  overflow: hidden;
  word-break: break-word;
}

.sponsored-card__caption {
  padding: 8px 10px 10px;
}

.sponsored-card__title {
  margin: 0;
  font-size: 13px;
  line-height: 1.4;
  font-weight: 600;
  color: var(--text-primary);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.sponsored-card__meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
  margin-top: 6px;
}

.sponsored-card__advertiser {
  flex: 1;
  min-width: 0;
  font-size: 11px;
  line-height: 1.3;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sponsored-card__ad-badge {
  flex-shrink: 0;
  font-size: 10px;
  line-height: 1;
  font-weight: 600;
  letter-spacing: 0.06em;
  color: #909090;
  border: 1px solid #d0d0d0;
  border-radius: 2px;
  padding: 2px 4px;
  user-select: none;
}
</style>
