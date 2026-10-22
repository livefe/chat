<template>
  <div v-if="userId > 0" class="user-cell">
    <el-link
      v-if="can(PERM.UserAccountDetail)"
      type="primary"
      :underline="false"
      class="user-cell__name"
      @click.stop="drawer.open(userId)"
    >
      {{ nickname || '未设置昵称' }}
    </el-link>
    <span v-else class="user-cell__name">{{ nickname || '未设置昵称' }}</span>
    <div class="sub">
      ID <CopyText :text="userId" />
      <template v-if="mobile"> · {{ mobile }}</template>
    </div>
  </div>
  <span v-else class="sub">-</span>
</template>

<script setup lang="ts">
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'
import { useUserDrawerStore } from '@/stores/userDrawer'
import CopyText from './CopyText.vue'

defineProps<{
  userId: number
  nickname?: string
  mobile?: string
}>()

const { can } = usePermission()
const drawer = useUserDrawerStore()
</script>

<style scoped>
.user-cell {
  line-height: 1.4;
  min-width: 0;
}

.user-cell__name {
  font-weight: 500;
}
</style>
