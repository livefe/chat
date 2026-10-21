<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { uploadImageMedia } from '@/api/upload'
import { updateProfile } from '@/api/user'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()

const avatarInput = ref<HTMLInputElement | null>(null)
const pendingAvatarUrl = ref('')
const saving = ref(false)
const uploadingAvatar = ref(false)

const nickname = computed(() => userStore.info?.nickname?.trim() || '')
const mobile = computed(() => userStore.info?.mobile?.trim() || '')
const avatarUrl = computed(() => pendingAvatarUrl.value || userStore.info?.avatarUrl || '')

async function persist(payload: { nickname?: string; avatarMediaId?: number }) {
  if (saving.value) return
  saving.value = true
  try {
    await updateProfile(payload)
    await userStore.fetchDetail()
    pendingAvatarUrl.value = ''
    showToast('保存成功')
  } finally {
    saving.value = false
  }
}

function pickAvatar() {
  if (uploadingAvatar.value || saving.value) return
  avatarInput.value?.click()
}

async function onAvatarSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  if (file.size > 10 * 1024 * 1024) {
    showToast('图片不能超过 10MB')
    return
  }

  uploadingAvatar.value = true
  try {
    const res = await uploadImageMedia(file)
    if (!res.ok) {
      if (res.message) showToast(res.message)
      return
    }
    pendingAvatarUrl.value = res.media.url
    await persist({ avatarMediaId: res.media.id })
  } finally {
    uploadingAvatar.value = false
  }
}

function openNicknameEditor() {
  if (saving.value) return
  router.push('/me/edit/nickname')
}

function openMobileEditor() {
  if (saving.value) return
  router.push('/me/edit/mobile')
}
</script>

<template>
  <SubPageLayout title="个人资料" full-height>
    <div class="profile-list">
      <van-cell
        title="头像"
        is-link
        center
        clickable
        :class="{ 'profile-list__cell--busy': uploadingAvatar || saving }"
        @click="pickAvatar"
      >
        <template #value>
          <UserAvatar :src="avatarUrl" :size="40" tone="profile" />
        </template>
      </van-cell>

      <van-cell
        title="昵称"
        is-link
        :value="nickname || '未设置'"
        clickable
        @click="openNicknameEditor"
      />

      <van-cell
        title="手机号"
        is-link
        :value="mobile || '未设置'"
        clickable
        @click="openMobileEditor"
      />
    </div>

    <input
      ref="avatarInput"
      class="profile-list__file-input"
      type="file"
      accept="image/*"
      @change="onAvatarSelected"
    />
  </SubPageLayout>
</template>

<style scoped>
.profile-list {
  margin-top: 12px;
  background: var(--card-bg);
}

.profile-list :deep(.van-cell) {
  padding: 14px 16px;
  font-size: 16px;
  color: var(--text-primary);
  background: var(--card-bg);
}

.profile-list :deep(.van-cell__title) {
  flex: none;
  width: 72px;
  margin-right: 12px;
}

.profile-list :deep(.van-cell__value) {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  color: var(--text-primary);
}

.profile-list :deep(.van-cell::after) {
  left: 16px;
  right: 0;
  border-color: var(--border-color-soft);
}

.profile-list__cell--busy {
  opacity: 0.6;
  pointer-events: none;
}

.profile-list__file-input {
  position: fixed;
  left: -9999px;
  width: 1px;
  height: 1px;
  opacity: 0;
}
</style>
