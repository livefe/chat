<template>
  <el-table
    :data="list"
    highlight-current-row
    empty-text="暂无下级区划"
    :row-class-name="rowClassName"
    @row-click="(row: RegionItem) => canViewDetail && emit('detail', row)"
  >
    <el-table-column prop="name" label="名称" min-width="160" />
    <el-table-column prop="shortName" label="简称" width="120" />
    <el-table-column label="层级" width="100">
      <template #default="{ row }">{{ regionLevelText(row.level) }}</template>
    </el-table-column>
    <el-table-column prop="code" label="区划码" width="160" />
    <el-table-column prop="fullPath" label="全路径" min-width="220" show-overflow-tooltip />
    <el-table-column label="操作" width="140" fixed="right">
      <template #default="{ row }">
        <div class="action-cell" @click.stop>
          <el-button
            v-if="row.level < REGION_LEAF_LEVEL"
            type="primary"
            link
            size="small"
            @click="emit('enter', row as RegionItem)"
          >
            进入下级
          </el-button>
          <el-button
            v-if="canViewDetail"
            type="primary"
            link
            size="small"
            @click="emit('detail', row as RegionItem)"
          >
            详情
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { RegionItem } from '@/api/region'
import { REGION_LEAF_LEVEL, regionLevelText } from '@/constants/region'

const props = defineProps<{
  list: RegionItem[]
  canViewDetail: boolean
  highlightCode?: string
}>()

const emit = defineEmits<{
  enter: [row: RegionItem]
  detail: [row: RegionItem]
}>()

function rowClassName({ row }: { row: RegionItem }) {
  return props.highlightCode && row.code === props.highlightCode ? 'is-located' : ''
}
</script>

<style scoped>
:deep(.el-table__row.is-located > td.el-table__cell) {
  background-color: var(--el-color-warning-light-9);
}
</style>
