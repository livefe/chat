<template>
  <el-input
    v-model="keyword"
    clearable
    placeholder="UID / 手机 / 昵称 / 会话ID"
    @keyup.enter="emit('search')"
  >
    <template #append>
      <el-button @click="emit('search')">搜索</el-button>
    </template>
  </el-input>
  <el-select
    v-if="showAssignee"
    v-model="assigneeId"
    clearable
    filterable
    placeholder="按坐席筛选"
    style="width: 100%; margin-top: 8px"
    @change="emit('change')"
  >
    <el-option
      v-for="c in agentOptions"
      :key="c.id"
      :label="`${c.name}（#${c.id}）`"
      :value="c.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import type { ImContactItem } from '@/api/im'

defineProps<{
  showAssignee: boolean
  agentOptions: ImContactItem[]
}>()

const keyword = defineModel<string>('keyword', { required: true })
const assigneeId = defineModel<number | undefined>('assigneeId')

const emit = defineEmits<{
  search: []
  change: []
}>()
</script>
