<script setup lang="ts">
import { computed, onActivated } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog } from 'vant'

import UserAvatar from '@/components/UserAvatar.vue'
import { useSupportStore } from '@/stores/support'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const supportStore = useSupportStore()
const router = useRouter()
const supportUnread = computed(() => supportStore.previewUnread)

onActivated(() => {
  void supportStore.refreshSession(true)
})

function logout() {
  showDialog({
    title: '退出登录',
    message: '确定退出吗？',
    showCancelButton: true,
  })
    .then(() => {
      userStore.logout()
      router.replace('/login')
    })
    .catch(() => {})
}
</script>

<template>
  <div class="me-page">
    <div class="me-profile-card" @click="router.push('/me/edit')">
      <UserAvatar :src="userStore.info?.avatarUrl" :size="56" tone="profile" />
      <div class="me-profile-card__info">
        <div class="me-profile-card__name">{{ userStore.info?.nickname || '' }}</div>
        <div v-if="userStore.info?.mobile" class="me-profile-card__meta">
          {{ userStore.info.mobile }}
        </div>
      </div>
      <van-icon name="arrow" class="me-profile-card__arrow" color="var(--text-secondary)" />
    </div>

    <van-cell-group inset class="me-section">
      <van-cell
        title="我发布的"
        is-link
        @click="router.push({ path: '/post/my', query: { mode: 'published' } })"
      />
      <van-cell
        title="我参与的"
        is-link
        @click="router.push({ path: '/post/my', query: { mode: 'joined' } })"
      />
    </van-cell-group>

    <van-cell-group inset class="me-section">
      <van-cell title="意见反馈" is-link @click="router.push('/me/feedback')" />
      <van-cell title="联系客服" is-link @click="router.push({ name: 'SupportChat' })">
        <template #right-icon>
          <span class="me-cell-end">
            <span v-if="supportUnread > 0" class="me-unread">
              {{ supportUnread > 99 ? '99+' : supportUnread }}
            </span>
            <van-icon name="arrow" />
          </span>
        </template>
      </van-cell>
    </van-cell-group>

    <van-cell-group inset class="me-section">
      <van-cell title="下载 App" is-link @click="router.push('/download')" />
    </van-cell-group>

    <van-cell-group inset class="me-section">
      <van-cell title="设置" is-link @click="router.push('/me/settings')" />
    </van-cell-group>

    <van-button class="page-btn" block @click="logout">退出登录</van-button>
  </div>
</template>

<style scoped>
.me-page {
  height: 100%;
  padding: 16px 0;
  background: var(--page-bg);
  overflow-y: auto;
}

.me-section {
  margin-top: 12px;
}

.me-cell-end {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: var(--icon-cell, 16px);
  color: var(--van-cell-right-icon-color);
  font-size: var(--icon-cell, 16px);
}

.me-unread {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  color: #fff;
  font-size: 10px;
  font-weight: 500;
  line-height: 1;
  background: var(--van-danger-color, #ee0a24);
  border-radius: 999px;
}
</style>
