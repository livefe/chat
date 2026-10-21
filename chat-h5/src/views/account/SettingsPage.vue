<script setup lang="ts">
import { useRouter } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'

const router = useRouter()
</script>

<template>
  <SubPageLayout title="设置">
    <van-cell-group inset class="settings-section">
      <van-cell title="黑名单" is-link @click="router.push('/me/settings/blocks')" />
      <van-cell title="注销账号" is-link @click="router.push('/me/account/deactivate')" />
    </van-cell-group>
  </SubPageLayout>
</template>

<style scoped>
.settings-section {
  margin-top: 12px;
}
</style>
