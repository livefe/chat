<script setup lang="ts">
import { useRouter } from 'vue-router'

import { goBackOr } from '@/utils/nav'

const props = withDefaults(
  defineProps<{
    title?: string
    fullHeight?: boolean
    wideTitle?: boolean
    onBack?: () => void
  }>(),
  { title: '' },
)

const router = useRouter()

function handleBack() {
  if (props.onBack) {
    props.onBack()
    return
  }
  goBackOr(router)
}
</script>

<template>
  <div class="sub-page" :class="{ 'sub-page--full': fullHeight }">
    <van-nav-bar
      :class="{ 'sub-page__nav--wide-title': wideTitle }"
      :title="$slots.title ? undefined : title || undefined"
      left-arrow
      :border="false"
      safe-area-inset-top
      @click-left="handleBack"
    >
      <template v-if="$slots.title" #title>
        <slot name="title" />
      </template>
      <template v-if="$slots.right" #right>
        <slot name="right" />
      </template>
    </van-nav-bar>
    <div class="sub-page__body">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.sub-page :deep(.van-nav-bar) {
  background: var(--page-bg);
}

.sub-page :deep(.van-nav-bar__left),
.sub-page :deep(.van-nav-bar__right) {
  position: relative;
  top: auto;
  bottom: auto;
}

.sub-page :deep(.van-nav-bar__title) {
  flex: 1;
  max-width: none;
  margin: 0;
  text-align: left;
  color: var(--text-primary);
  font-size: var(--font-page-title);
  font-weight: 600;
}
</style>
