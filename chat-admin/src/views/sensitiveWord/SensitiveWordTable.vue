<template>
  <el-table
    ref="tableRef"
    :data="list"
    v-loading="loading"
    row-key="id"
    @selection-change="(rows: SensitiveWordItem[]) => emit('selection-change', rows)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无屏蔽词" @retry="emit('page-change')" />
    </template>
    <el-table-column v-if="selectable" type="selection" width="44" />
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column prop="word" label="屏蔽词" min-width="140" />
    <el-table-column prop="categoryText" label="分类" width="100" />
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
    <el-table-column prop="updatedAt" label="更新时间" width="170" />
    <el-table-column label="操作" width="160" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.OpsSensitiveWordUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as SensitiveWordItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="can(PERM.OpsSensitiveWordStatus)"
            size="small"
            link
            :type="row.status === 1 ? 'warning' : 'success'"
            @click="emit('toggle-status', row as SensitiveWordItem)"
          >
            {{ row.status === 1 ? '停用' : '启用' }}
          </el-button>
          <el-button
            v-if="can(PERM.OpsSensitiveWordDelete)"
            size="small"
            link
            type="danger"
            @click="emit('delete', row as SensitiveWordItem)"
          >
            删除
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div class="pagination">
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="PAGE_SIZES"
      :layout="PAGINATION_LAYOUT"
      @size-change="emit('page-change')"
      @current-change="emit('page-change')"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { TableInstance } from 'element-plus'
import type { SensitiveWordItem } from '@/api/sensitiveWord'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'

defineProps<{
  list: SensitiveWordItem[]
  loading: boolean
  error?: boolean
  total: number
  selectable: boolean
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  'selection-change': [rows: SensitiveWordItem[]]
  edit: [row: SensitiveWordItem]
  'toggle-status': [row: SensitiveWordItem]
  delete: [row: SensitiveWordItem]
  'page-change': []
}>()

const { can } = usePermission()
const tableRef = ref<TableInstance>()

defineExpose({
  clearSelection: () => tableRef.value?.clearSelection(),
})
</script>
