<script setup lang="ts">
import { ref, watch } from 'vue'
import { showFailToast } from 'vant'

import type { PostSignupItem } from '@/api/post'

const REMARK_MAX = 100

const props = defineProps<{
  target: PostSignupItem | null
  save: (item: PostSignupItem, remark: string) => Promise<boolean>
}>()

const emit = defineEmits<{
  close: []
}>()

const visible = ref(false)
const saving = ref(false)
const draft = ref('')

watch(
  () => props.target,
  (item) => {
    if (!item) return
    draft.value = (item.remark || '').trim()
    visible.value = true
  },
)

function close() {
  if (saving.value) return
  visible.value = false
  draft.value = ''
  emit('close')
}

async function onSave() {
  const target = props.target
  if (!target || saving.value) return
  const remark = draft.value.trim()
  if ([...remark].length > REMARK_MAX) {
    showFailToast(`备注不能超过${REMARK_MAX}字`)
    return
  }
  saving.value = true
  let ok = false
  try {
    ok = await props.save(target, remark)
  } finally {
    saving.value = false
  }
  if (ok) close()
}
</script>

<template>
  <van-popup
    v-model:show="visible"
    position="bottom"
    round
    safe-area-inset-bottom
    :close-on-click-overlay="!saving"
    @closed="close"
  >
    <div class="remark-sheet">
      <div class="remark-sheet__head">
        <span class="remark-sheet__title">备注</span>
        <span class="remark-sheet__name">{{ target?.userNickname }}</span>
      </div>
      <van-field
        v-model="draft"
        type="textarea"
        rows="3"
        autosize
        :maxlength="REMARK_MAX"
        show-word-limit
        placeholder="例如：已交费、特殊需求等"
      />
      <div class="remark-sheet__actions">
        <button
          type="button"
          class="remark-sheet__btn remark-sheet__btn--ghost"
          :disabled="saving"
          @click="close"
        >
          取消
        </button>
        <button
          type="button"
          class="remark-sheet__btn remark-sheet__btn--primary"
          :disabled="saving"
          @click="onSave"
        >
          {{ saving ? '保存中…' : '确定' }}
        </button>
      </div>
    </div>
  </van-popup>
</template>

<style scoped>
.remark-sheet {
  padding: 16px 16px calc(16px + env(safe-area-inset-bottom));
}

.remark-sheet__head {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 8px;
  padding: 0 4px;
}

.remark-sheet__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.remark-sheet__name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  color: var(--text-tertiary);
}

.remark-sheet__actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.remark-sheet__btn {
  flex: 1;
  height: 40px;
  border-radius: 999px;
  border: none;
  font-size: 15px;
  font-weight: 500;
}

.remark-sheet__btn--ghost {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-primary);
}

.remark-sheet__btn--primary {
  background: var(--brand-color);
  color: #fff;
}

.remark-sheet__btn:disabled {
  opacity: 0.6;
}
</style>
