<script setup lang="ts">
import { watch } from 'vue'
import { useRoute } from 'vue-router'
import { ConfigProvider } from 'vant'

import { useAppViewport } from '@/composables/useAppViewport'
import { useChatNoticeToasts } from '@/composables/useChatNoticeToasts'
import TextEditMenu from '@/components/TextEditMenu.vue'
import { isInAppShell } from '@/utils/appShell'

useAppViewport()
useChatNoticeToasts()

const route = useRoute()

watch(
  () => route.query.from,
  () => {
    if (isInAppShell(route.query)) {
      document.documentElement.dataset.appShell = '1'
    } else {
      delete document.documentElement.dataset.appShell
    }
  },
  { immediate: true },
)
</script>

<template>
  <ConfigProvider
    :theme-vars="{
      primaryColor: 'var(--brand-color)',
      dangerColor: 'var(--text-danger)',
      navBarBackground: 'var(--page-bg)',
      navBarHeight: '56px',
      navBarTitleFontSize: 'var(--font-page-title)',
      navBarArrowSize: 'var(--nav-back-icon-size)',
      navBarIconColor: 'var(--text-primary)',
      navBarTextColor: 'var(--text-primary)',
      cellGroupBackground: 'var(--card-bg)',
    }"
  >
    <router-view />
    <TextEditMenu />
  </ConfigProvider>
</template>
