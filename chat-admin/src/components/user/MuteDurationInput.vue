<template>
  <div class="mute-duration">
    <el-radio-group v-model="mode">
      <el-radio-button v-for="p in MUTE_PRESETS" :key="p.value" :value="p.value">
        {{ p.label }}
      </el-radio-button>
      <el-radio-button :value="CUSTOM">自定义</el-radio-button>
    </el-radio-group>
    <div v-if="mode === CUSTOM" class="mute-duration__custom">
      <el-input-number v-model="customHours" :min="1" :max="MUTE_MAX_HOURS" :step="1" />
      <span class="sub">小时</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

import { MUTE_MAX_HOURS, MUTE_PRESETS } from '@/constants/user'

/** 0 在禁言语义里是「解除」，这里借用它表示自定义。 */
const CUSTOM = 0

const hours = defineModel<number>({ required: true })

const isPreset = (v: number) => MUTE_PRESETS.some((p) => p.value === v)
const mode = ref<number>(isPreset(hours.value) ? hours.value : CUSTOM)
const customHours = ref(isPreset(hours.value) || hours.value <= 0 ? 12 : hours.value)

watch([mode, customHours], () => {
  hours.value = mode.value === CUSTOM ? customHours.value : mode.value
})
</script>

<style scoped>
.mute-duration {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.mute-duration__custom {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
