<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { updateProfile } from '@/api/user'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()

const original = ref(userStore.info?.nickname || '')
const draft = ref(userStore.info?.nickname || '')
const saving = ref(false)

const canSave = computed(() => {
  const next = draft.value.trim()
  return !!next && next !== original.value.trim() && !saving.value
})

onMounted(() => {
  const name = userStore.info?.nickname || ''
  original.value = name
  draft.value = name
})

async function save() {
  if (!canSave.value) return
  const next = draft.value.trim()
  saving.value = true
  try {
    await updateProfile({ nickname: next })
    await userStore.fetchDetail()
    showToast('保存成功')
    router.back()
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <SubPageLayout title="更改昵称" full-height class="nickname-edit">
    <template #right>
      <button
        type="button"
        class="nickname-edit__save"
        :class="{ 'nickname-edit__save--active': canSave }"
        :disabled="!canSave"
        @click="save"
      >
        {{ saving ? '保存中' : '保存' }}
      </button>
    </template>

    <div class="nickname-edit__body">
      <div class="nickname-edit__field">
        <input
          v-model="draft"
          class="nickname-edit__input"
          maxlength="32"
          type="text"
          autofocus
          placeholder="请输入昵称"
        />
      </div>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.nickname-edit :deep(.van-nav-bar) {
  background: var(--card-bg);
}

.nickname-edit__save {
  min-width: 56px;
  height: 32px;
  padding: 0 12px;
  border: none;
  border-radius: 4px;
  background: var(--brand-color);
  color: #fff;
  font-size: 15px;
  line-height: 32px;
  opacity: 0.35;
}

.nickname-edit__save--active {
  opacity: 1;
}

.nickname-edit__save:disabled {
  opacity: 0.35;
}

.nickname-edit__body {
  padding: 28px 20px 0;
}

.nickname-edit__field {
  border-bottom: 1px solid var(--border-color-soft);
}

.nickname-edit__input {
  width: 100%;
  height: 44px;
  padding: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 17px;
  line-height: 44px;
}

.nickname-edit__input::placeholder {
  color: var(--text-tertiary);
}
</style>
