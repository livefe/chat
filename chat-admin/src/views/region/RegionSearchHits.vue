<template>
  <div class="tree-manage-hits">
    <div class="tree-manage-hits__head">
      <span>搜索结果（点击定位）</span>
      <el-button link type="primary" @click="emit('close')">收起</el-button>
    </div>
    <el-table
      :data="hits"
      v-loading="loading"
      size="small"
      max-height="180"
      highlight-current-row
      empty-text="暂无匹配"
      @row-click="(row: RegionItem) => emit('locate', row)"
    >
      <el-table-column prop="name" label="名称" min-width="120" />
      <el-table-column prop="fullPath" label="全路径" min-width="240" show-overflow-tooltip />
      <el-table-column label="层级" width="90">
        <template #default="{ row }">{{ regionLevelText(row.level) }}</template>
      </el-table-column>
      <el-table-column prop="code" label="区划码" width="140" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import type { RegionItem } from '@/api/region'
import { regionLevelText } from '@/constants/region'

defineProps<{
  hits: RegionItem[]
  loading: boolean
}>()

const emit = defineEmits<{
  locate: [row: RegionItem]
  close: []
}>()
</script>
