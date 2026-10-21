<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无审计日志">
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column label="操作人" min-width="140">
      <template #default="{ row }">
        <div>{{ row.employeeName || '-' }}</div>
        <div class="sub">{{ row.employeeMobile }} · ID {{ row.employeeId }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="moduleText" label="模块" width="110" />
    <el-table-column prop="actionText" label="操作" width="130" />
    <el-table-column label="结果" width="80">
      <template #default="{ row }">
        <el-tag :type="row.success ? 'success' : 'danger'" size="small">
          {{ row.successText || (row.success ? '成功' : '失败') }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="目标" min-width="160">
      <template #default="{ row }">
        <div>{{ row.targetLabel || '-' }}</div>
        <div v-if="row.targetType" class="sub">{{ row.targetType }}</div>
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
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'

defineProps<{
  list: AuditLogItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [row: AuditLogItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>
