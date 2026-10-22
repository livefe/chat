<template>
  <el-table :data="list" v-loading="loading">
    <template #empty>
      <TableEmpty :error="error" text="暂无审计记录" @retry="emit('page-change')" />
    </template>
    <el-table-column label="操作人" min-width="150">
      <template #default="{ row }">
        <el-link type="primary" :underline="false" @click="emit('filter-employee', row.employeeId)">
          {{ row.employeeName || '-' }}
        </el-link>
        <div class="sub">{{ row.employeeMobile }} · ID {{ row.employeeId }}</div>
      </template>
    </el-table-column>
    <el-table-column label="操作" min-width="160">
      <template #default="{ row }">
        <div>{{ row.actionText }}</div>
        <div class="sub">{{ row.moduleText }}</div>
      </template>
    </el-table-column>
    <el-table-column label="结果" width="90">
      <template #default="{ row }">
        <el-tooltip :disabled="row.success || !row.resultMessage" :content="row.resultMessage">
          <el-tag :type="row.success ? 'success' : 'danger'" size="small">
            {{ row.successText }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column label="对象" min-width="200">
      <template #default="{ row }">
        <div>{{ row.targetLabel || '-' }}</div>
        <div v-if="row.targetTypeText" class="sub">{{ row.targetTypeText }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="ip" label="IP" width="130" />
    <el-table-column prop="createdAt" label="操作时间" width="170" />
    <el-table-column label="操作" width="80" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.OpsAuditLogDetail)"
            size="small"
            link
            type="primary"
            @click="emit('detail', row as AuditLogItem)"
          >
            详情
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
import type { AuditLogItem } from '@/api/auditLog'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'

defineProps<{
  list: AuditLogItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [row: AuditLogItem]
  'filter-employee': [employeeId: number]
  'page-change': []
}>()

const { can } = usePermission()
</script>
