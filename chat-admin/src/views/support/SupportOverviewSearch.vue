<template>
  <el-input
    v-model="keyword"
    clearable
    placeholder="UID / 手机 / 昵称 / 会话ID"
    @keyup.enter="emit('search')"
    @clear="emit('search')"
  >
    <template #append>
      <el-button @click="emit('search')">搜索</el-button>
    </template>
  </el-input>
  <template v-if="showAdvanced">
    <el-select
      v-model="assigneeId"
      clearable
      filterable
      placeholder="按坐席筛选"
      class="overview-search__field"
      @change="emit('search')"
    >
      <el-option v-for="a in agentOptions" :key="a.id" :label="a.name" :value="a.id">
        <span>{{ a.name }}（#{{ a.id }}）</span>
        <span class="overview-search__sub">{{ a.online ? '在线' : '离线' }}</span>
      </el-option>
    </el-select>
    <el-date-picker
      v-model="dateRange"
      type="daterange"
      value-format="YYYY-MM-DD"
      :start-placeholder="`${dateLabel}起`"
      end-placeholder="止"
      unlink-panels
      class="overview-search__field"
      @change="emit('search')"
    />
  </template>
</template>

<script setup lang="ts">
import type { SupportAgent } from '@/api/support'

defineProps<{
  showAdvanced: boolean
  dateLabel: string
  agentOptions: SupportAgent[]
}>()

const keyword = defineModel<string>('keyword', { required: true })
const assigneeId = defineModel<number | undefined>('assigneeId')
const dateRange = defineModel<[string, string] | null>('dateRange', { required: true })

const emit = defineEmits<{ search: [] }>()
</script>

<style scoped>
.overview-search__field {
  width: 100%;
  margin-top: 8px;
}

.overview-search__field:deep(.el-range-input) {
  width: 42%;
}

.overview-search__sub {
  float: right;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
