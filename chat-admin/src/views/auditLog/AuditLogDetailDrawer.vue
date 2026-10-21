<template>
  <el-drawer v-model="visible" title="审计详情" size="640px">
    <div v-loading="loading">
      <template v-if="detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="记录ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="操作人">
            {{ detail.employeeName || '-' }}（{{ detail.employeeMobile }} / ID
            {{ detail.employeeId }}）
          </el-descriptions-item>
          <el-descriptions-item label="模块">{{ detail.moduleText }}</el-descriptions-item>
          <el-descriptions-item label="操作">{{ detail.actionText }}</el-descriptions-item>
          <el-descriptions-item label="结果">
            <el-tag :type="detail.success ? 'success' : 'danger'" size="small">
              {{ detail.successText || (detail.success ? '成功' : '失败') }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="目标">
            {{ detail.targetLabel || '-' }}
            <span v-if="detail.targetType">（{{ detail.targetType }}）</span>
          </el-descriptions-item>
          <el-descriptions-item label="请求"
            >{{ detail.requestMethod }} {{ detail.requestPath }}</el-descriptions-item
          >
          <el-descriptions-item label="IP">{{ detail.ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="UA">{{ detail.userAgent || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作时间">{{ detail.createdAt }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">请求详情</el-divider>
        <pre class="detail-json">{{ formatDetail(detail.detail) }}</pre>
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getAuditLogDetail, type AuditLogItem } from '@/api/auditLog'
import { showApiError } from '@/utils/apiError'

const visible = ref(false)
const loading = ref(false)
const detail = ref<AuditLogItem | null>(null)

async function open(row: AuditLogItem) {
  visible.value = true
  loading.value = true
  detail.value = null
  try {
    const res = await getAuditLogDetail(row.id)
    detail.value = res.data
  } catch (e) {
    showApiError(e, '获取审计详情失败')
    visible.value = false
  } finally {
    loading.value = false
  }
}

function formatDetail(raw: string) {
  if (!raw) return '-'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

defineExpose({ open })
</script>

<style scoped>
.detail-json {
  margin: 0;
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  line-height: 1.6;
}
</style>
