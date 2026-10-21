<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { PickerOption } from 'vant'

import { formatDateTimeChinese, splitDateTime, toStorageDateTime } from '@/utils/datetime'

withDefaults(
  defineProps<{
    label: string
    placeholder?: string
    title?: string
  }>(),
  {
    placeholder: '请选择',
    title: '选择时间',
  },
)

const model = defineModel<string>({ default: '' })

const show = ref(false)
const pickerDate = ref<string[]>([])
const pickerTime = ref<string[]>([])

const hasValue = computed(() => !!formatDateTimeChinese(model.value))
const displayText = computed(() => formatDateTimeChinese(model.value))

function dateFormatter(type: string, option: PickerOption) {
  if (type === 'year') option.text = `${option.text}年`
  if (type === 'month') option.text = `${option.text}月`
  if (type === 'day') option.text = `${option.text}日`
  return option
}

function timeFormatter(type: string, option: PickerOption) {
  if (type === 'hour') option.text = `${option.text}时`
  if (type === 'minute') option.text = `${option.text}分`
  return option
}

function applyFromPicker() {
  const [y, m, d] = pickerDate.value
  const [hh, mm] = pickerTime.value
  if (!y || !m || !d || hh === undefined || hh === '' || mm === undefined || mm === '') return
  model.value = toStorageDateTime(String(y), String(m), String(d), String(hh), String(mm))
}

function openPicker() {
  const active = document.activeElement
  if (active instanceof HTMLElement) active.blur()
  const { date, time } = splitDateTime(model.value)
  pickerDate.value = date
  pickerTime.value = time
  show.value = true
  applyFromPicker()
}

function onReadonlyFocus(e: FocusEvent) {
  const t = e.target
  if (t instanceof HTMLElement) t.blur()
}

function confirmPick() {
  applyFromPicker()
  show.value = false
}

function cancelPick() {
  applyFromPicker()
  show.value = false
}

watch([pickerDate, pickerTime], () => {
  if (!show.value) return
  applyFromPicker()
})
</script>

<template>
  <van-field :label="label" readonly is-link clickable @click="openPicker" @focus="onReadonlyFocus">
    <template #input>
      <span :class="hasValue ? 'dt-value' : 'dt-placeholder'">
        {{ hasValue ? displayText : placeholder }}
      </span>
    </template>
  </van-field>

  <van-popup
    v-model:show="show"
    position="bottom"
    round
    safe-area-inset-bottom
    teleport="body"
    :lock-scroll="false"
  >
    <div class="app-sheet dt-panel" @touchmove.stop>
      <div class="app-sheet__handle" />
      <div class="dt-toolbar">
        <button type="button" class="dt-action" @click="cancelPick">取消</button>
        <span class="dt-title">{{ title }}</span>
        <button type="button" class="dt-action dt-action--primary" @click="confirmPick">
          确定
        </button>
      </div>
      <div class="dt-body">
        <div class="dt-col dt-col--date">
          <van-date-picker
            v-model="pickerDate"
            :show-toolbar="false"
            :formatter="dateFormatter"
            :visible-option-num="5"
          />
        </div>
        <div class="dt-col dt-col--time">
          <van-time-picker
            v-model="pickerTime"
            :show-toolbar="false"
            :formatter="timeFormatter"
            :columns-type="['hour', 'minute']"
            :visible-option-num="5"
          />
        </div>
      </div>
    </div>
  </van-popup>
</template>

<style scoped>
:deep(.van-field__control--custom) {
  width: 100%;
  justify-content: flex-end;
}
.dt-placeholder,
.dt-value {
  display: block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
}
.dt-placeholder {
  color: var(--text-tertiary);
}
.dt-value {
  color: var(--text-primary);
}

.dt-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px 12px;
  border-bottom: 1px solid var(--border-color);
}

.dt-title {
  font-size: var(--sheet-title-size);
  font-weight: var(--sheet-title-weight);
  color: var(--text-primary);
}

.dt-action {
  border: none;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 15px;
  padding: 0;
}

.dt-action--primary {
  color: var(--brand-color-dark);
  font-weight: 600;
}

.dt-body {
  display: flex;
  align-items: stretch;
}

.dt-col {
  flex: 1;
  overflow: hidden;
}

.dt-col--date {
  flex: 3;
}

.dt-col--time {
  flex: 2;
  border-left: 1px solid #eeeeee;
}

.dt-col :deep(.van-picker) {
  width: 100%;
}

.dt-col :deep(.van-picker__columns) {
  width: 100%;
}

.dt-col :deep(.van-picker-column) {
  flex: 1;
  min-width: 0;
}
</style>
