<script setup lang="ts">
import type { PostDraftForm } from '@/stores/postDraft'
import { formatOptionalNumber, parseHelpReward } from '@/utils/post/postPublishPayload'

const form = defineModel<PostDraftForm>('form', { required: true })

const emit = defineEmits<{
  'blur-input': []
}>()

function onHelpRewardChange() {
  emit('blur-input')
}

function onHelpRewardAmountInput(value: string) {
  parseHelpReward(form.value, value)
}
</script>

<template>
  <div class="form-card">
    <van-cell title="提供报酬">
      <template #right-icon>
        <van-switch v-model="form.helpReward" size="20px" @change="onHelpRewardChange" />
      </template>
    </van-cell>
    <template v-if="form.helpReward">
      <van-field
        :model-value="formatOptionalNumber(form.helpRewardAmount)"
        label="报酬金额"
        type="number"
        input-align="right"
        placeholder="请输入金额"
        @update:model-value="onHelpRewardAmountInput"
      >
        <template #extra>元</template>
      </van-field>
      <van-field
        v-model="form.helpRewardDesc"
        label="报酬说明"
        input-align="right"
        placeholder="选填"
      />
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
