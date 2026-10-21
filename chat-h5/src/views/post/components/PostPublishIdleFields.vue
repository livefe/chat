<script setup lang="ts">
import type { PostDraftForm } from '@/stores/postDraft'
import { formatOptionalNumber, parseIdlePrice } from '@/utils/post/postPublishPayload'

const form = defineModel<PostDraftForm>('form', { required: true })

const emit = defineEmits<{
  'blur-input': []
}>()

function onEnableIdlePriceChange() {
  emit('blur-input')
}

function onIdlePriceInput(value: string) {
  parseIdlePrice(form.value, value)
}
</script>

<template>
  <div class="form-card">
    <van-cell title="金额设置">
      <template #right-icon>
        <van-switch v-model="form.enableIdlePrice" size="20px" @change="onEnableIdlePriceChange" />
      </template>
    </van-cell>
    <template v-if="form.enableIdlePrice">
      <van-field
        :model-value="formatOptionalNumber(form.idlePrice)"
        label="价格"
        type="number"
        input-align="right"
        placeholder="请输入价格"
        @update:model-value="onIdlePriceInput"
      >
        <template #extra>元</template>
      </van-field>
      <van-cell title="可议价">
        <template #right-icon>
          <van-switch v-model="form.idleNegotiable" size="20px" />
        </template>
      </van-cell>
    </template>
  </div>
</template>

<style scoped>
.form-card {
  overflow: clip;
  border-radius: 16px;
  background: var(--card-bg);
}
.form-card :deep(.van-cell) {
  background: transparent;
}
.form-card :deep(.van-cell::after) {
  left: 16px;
  right: 0;
  border-color: var(--border-color);
}
.form-card :deep(.van-cell:not(:has(~ .van-cell))::after) {
  display: none;
}
.form-card :deep(.van-field__label) {
  width: 5.5em;
  flex: none;
  margin-right: 16px;
  color: var(--text-primary);
  white-space: nowrap;
}
.form-card :deep(.van-field__value),
.form-card :deep(.van-field__control) {
  min-width: 0;
}
</style>
