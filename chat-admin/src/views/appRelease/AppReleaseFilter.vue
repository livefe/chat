<template>
  <el-select
    v-model="platform"
    clearable
    placeholder="全部平台"
    style="width: 140px"
    @change="emit('search')"
  >
    <el-option
      v-for="opt in APP_PLATFORM_OPTIONS"
      :key="opt.value"
      :label="opt.label"
      :value="opt.value"
    />
  </el-select>
</template>

<script setup lang="ts">
import { APP_PLATFORM_OPTIONS } from '@/constants/appRelease'

const platform = defineModel<string>('platform', { required: true })
const emit = defineEmits<{ search: [] }>()
</script>
