<template>
  <section class="review-section">
    <div class="section-title">被举报用户现状</div>
    <div class="user-card">
      <el-avatar :src="user.avatarUrl || undefined" :size="44">
        {{ (user.nickname || '用').slice(0, 1) }}
      </el-avatar>
      <div class="user-card__main">
        <div class="user-card__name">
          <UserCell :user-id="user.userId" :nickname="user.nickname" :mobile="user.mobile" />
        </div>
        <div class="user-card__tags">
          <el-tag v-if="user.deactivated" type="info" size="small">已注销</el-tag>
          <el-tag v-else :type="user.status === 1 ? 'success' : 'danger'" size="small">
            {{ user.statusText }}
          </el-tag>
          <el-tag v-if="user.muteText" type="warning" size="small">{{ user.muteText }}</el-tag>
        </div>
      </div>
      <div class="user-card__stats">
        <div class="stat">
          <el-link type="primary" :underline="false" @click="emit('show-reports')">
            {{ user.reportCount }}
          </el-link>
          <span class="sub">累计被举报</span>
        </div>
        <div class="stat">
          <el-link
            v-if="can(PERM.UserAccountDetail)"
            type="primary"
            :underline="false"
            @click="drawer.open(user.userId, 'punishments')"
          >
            {{ user.punishmentCount }}
          </el-link>
          <span v-else>{{ user.punishmentCount }}</span>
          <span class="sub">处罚记录</span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { UserGovernanceSummary } from '@/api/user'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { useUserDrawerStore } from '@/stores/userDrawer'

defineProps<{ user: UserGovernanceSummary }>()
const emit = defineEmits<{ 'show-reports': [] }>()

const { can } = usePermission()
const drawer = useUserDrawerStore()
</script>

<style scoped>
.user-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.user-card__main {
  flex: 1;
  min-width: 0;
}

.user-card__tags {
  display: flex;
  gap: 6px;
  margin-top: 4px;
}

.user-card__stats {
  display: flex;
  gap: 20px;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
}

.stat .sub {
  font-size: 12px;
  font-weight: normal;
}
</style>
