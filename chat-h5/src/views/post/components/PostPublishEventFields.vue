<script setup lang="ts">
import { computed } from 'vue'

import DateTimeField from './DateTimeField.vue'
import type { PostDraftForm } from '@/stores/postDraft'
import {
  formatOptionalNumber,
  parseFeeAmount,
  parseMaxParticipants,
} from '@/utils/post/postPublishPayload'

const form = defineModel<PostDraftForm>('form', { required: true })

const emit = defineEmits<{
  'blur-input': []
}>()

const isCharge = computed(() => form.value.feeType === 'charge')

function onReadonlyFieldFocus(e: FocusEvent) {
  const t = e.target
  if (t instanceof HTMLElement) t.blur()
}

function onEnableScheduleChange() {
  emit('blur-input')
}

function onEnableFeeChange() {
  emit('blur-input')
}

function onNeedSignupChange() {
  emit('blur-input')
  if (!form.value.needSignup) {
    form.value.needCheckIn = false
  } else if (!form.value.enableSchedule) {
    form.value.enableSchedule = true
  }
}

function onLimitParticipantsChange() {
  emit('blur-input')
}

function onFeeAmountInput(value: string) {
  parseFeeAmount(form.value, value)
}

function onMaxParticipantsInput(value: string) {
  parseMaxParticipants(form.value, value)
}
</script>

<template>
  <div class="form-card">
    <van-cell title="时间设置">
      <template #right-icon>
        <van-switch v-model="form.enableSchedule" size="20px" @change="onEnableScheduleChange" />
      </template>
    </van-cell>
    <template v-if="form.enableSchedule">
      <DateTimeField v-model="form.startAt" label="开始时间" title="选择开始时间" />
      <DateTimeField v-model="form.endAt" label="结束时间" title="选择结束时间" />
    </template>
    <van-cell title="金额设置">
      <template #right-icon>
        <van-switch v-model="form.enableFee" size="20px" @change="onEnableFeeChange" />
      </template>
    </van-cell>
    <template v-if="form.enableFee">
      <van-field label="类型" readonly @focus="onReadonlyFieldFocus">
        <template #input>
          <van-radio-group v-model="form.feeType" direction="horizontal" class="fee-type-group">
            <van-radio name="charge">收费</van-radio>
            <van-radio name="reward">报酬</van-radio>
          </van-radio-group>
        </template>
      </van-field>
      <van-field
        :model-value="formatOptionalNumber(form.feeAmount)"
        :label="isCharge ? '收费金额' : '报酬金额'"
        type="number"
        input-align="right"
        placeholder="请输入金额"
        @update:model-value="onFeeAmountInput"
      >
        <template #extra>元</template>
      </van-field>
      <van-field
        v-model="form.feeDescription"
        :label="isCharge ? '收费说明' : '报酬说明'"
        input-align="right"
        placeholder="选填"
      />
    </template>
    <van-cell title="需要报名">
      <template #right-icon>
        <van-switch v-model="form.needSignup" size="20px" @change="onNeedSignupChange" />
      </template>
    </van-cell>
    <template v-if="form.needSignup">
      <DateTimeField v-model="form.signupDeadlineAt" label="报名截止" title="选择报名截止" />
      <van-cell title="限制人数">
        <template #right-icon>
          <van-switch
            v-model="form.limitParticipants"
            size="20px"
            @change="onLimitParticipantsChange"
          />
        </template>
      </van-cell>
      <van-field
        v-if="form.limitParticipants"
        :model-value="formatOptionalNumber(form.maxParticipants)"
        label="人数上限"
        type="digit"
        input-align="right"
        placeholder="请输入人数"
        @update:model-value="onMaxParticipantsInput"
      >
        <template #extra>人</template>
      </van-field>
      <van-cell v-if="form.limitParticipants" title="开启候补">
        <template #right-icon>
          <van-switch v-model="form.allowWaitlist" size="20px" />
        </template>
      </van-cell>
      <van-cell title="需要核验">
        <template #right-icon>
          <van-switch v-model="form.needCheckIn" size="20px" />
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
.form-card :deep(.van-field__control--custom) {
  overflow: hidden;
}
.fee-type-group {
  flex-wrap: wrap;
  gap: 8px 20px;
}
</style>
