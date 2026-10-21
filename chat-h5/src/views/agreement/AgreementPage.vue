<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { isInAppShell } from '@/utils/appShell'

import LegalDocument from './components/LegalDocument.vue'
import { privacyPolicy } from './content/privacyPolicy'
import { userAgreement } from './content/userAgreement'

const props = defineProps<{ doc: 'user' | 'privacy' }>()
const route = useRoute()
const inApp = computed(() => isInAppShell(route.query))

const content = computed(() => (props.doc === 'user' ? userAgreement : privacyPolicy))
const navTitle = computed(() => (props.doc === 'user' ? '用户协议' : '隐私政策'))

onMounted(() => {
  document.title = content.value.title
})
</script>

<template>
  <SubPageLayout v-if="!inApp" :title="navTitle">
    <div class="agreement-content">
      <LegalDocument :doc="content" />
    </div>
  </SubPageLayout>
  <div v-else class="agreement-app">
    <LegalDocument :doc="content" />
  </div>
</template>

<style scoped>
.agreement-app {
  min-height: 100vh;
  padding: 16px 16px calc(16px + env(safe-area-inset-bottom));
  background: #fff;
  box-sizing: border-box;
}
.agreement-content {
  margin: 12px;
  padding: 16px;
  border-radius: 12px;
  background: var(--card-bg);
}
</style>
