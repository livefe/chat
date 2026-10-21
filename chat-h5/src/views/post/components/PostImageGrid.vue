<script setup lang="ts">
import { computed, ref } from 'vue'
import { showImagePreview } from 'vant'

import { resolveMediaUrls } from '@/utils/media'

const props = defineProps<{
  images: string[]
}>()

const visibleImages = computed(() => resolveMediaUrls(props.images))

const swipeIndex = ref(0)

const swipeIndicator = computed(() => `${swipeIndex.value + 1}/${visibleImages.value.length}`)

function preview(index: number) {
  if (!visibleImages.value.length) return
  showImagePreview({
    images: visibleImages.value,
    startPosition: index,
  })
}

function onSwipeChange(index: number) {
  swipeIndex.value = index
}
</script>

<template>
  <div v-if="visibleImages.length" class="post-images">
    <button
      v-if="visibleImages.length === 1"
      type="button"
      class="detail-single"
      @click="preview(0)"
    >
      <img :src="visibleImages[0]" alt="" class="detail-single__img" loading="lazy" />
    </button>

    <div v-else class="detail-swipe-wrap">
      <van-swipe class="detail-swipe" :loop="false" indicator-color="#fff" @change="onSwipeChange">
        <van-swipe-item v-for="(url, index) in visibleImages" :key="`${url}-${index}`">
          <button type="button" class="detail-swipe__btn" @click="preview(index)">
            <img :src="url" alt="" class="detail-swipe__img" loading="lazy" />
          </button>
        </van-swipe-item>
      </van-swipe>
      <span class="detail-swipe__count">{{ swipeIndicator }}</span>
    </div>
  </div>
</template>

<style scoped>
.post-images {
  margin-bottom: 16px;
}

.detail-single {
  display: block;
  width: 100%;
  padding: 0;
  border: none;
  border-radius: 8px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.04);
}

.detail-single__img {
  display: block;
  width: 100%;
  height: auto;
  max-height: 320px;
  object-fit: cover;
}

.detail-swipe-wrap {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.04);
}

.detail-swipe {
  aspect-ratio: 4 / 3;
}

.detail-swipe :deep(.van-swipe__track),
.detail-swipe :deep(.van-swipe-item) {
  height: 100%;
}

.detail-swipe__btn {
  display: block;
  width: 100%;
  height: 100%;
  padding: 0;
  border: none;
  background: transparent;
}

.detail-swipe__img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-swipe__count {
  position: absolute;
  right: 10px;
  bottom: 10px;
  z-index: 1;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 12px;
  line-height: 1.5;
  pointer-events: none;
}
</style>
