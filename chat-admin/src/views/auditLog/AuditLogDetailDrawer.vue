<template>
  <el-drawer v-model="visible" title="审计详情" size="640px">
    <div v-loading="loading" class="audit-detail">
      <template v-if="detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="记录ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="操作人">
            {{ detail.employeeName || '-' }}（{{ detail.employeeMobile }} / ID
            {{ detail.employeeId }}）
          </el-descriptions-item>
          <el-descriptions-item label="模块">{{ detail.moduleText }}</el-descriptions-item>
          <el-descriptions-item label="操作">
            {{ detail.actionText }}
            <span class="sub">{{ detail.action }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="结果">
            <el-tag :type="detail.success ? 'success' : 'danger'" size="small">
              {{ detail.successText }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="!detail.success" label="失败原因">
            {{ detail.resultMessage || '-' }}
            <span v-if="detail.resultCode" class="sub">（{{ detail.resultCode }}）</span>
          </el-descriptions-item>
          <el-descriptions-item label="对象">
            <span>{{ detail.targetLabel || '-' }}</span>
            <span v-if="detail.targetTypeText" class="sub">（{{ detail.targetTypeText }}）</span>
            <el-button
              v-if="targetLink"
              link
              type="primary"
              class="target-link"
              @click="targetLink.open"
            >
              查看对象
            </el-button>
          </el-descriptions-item>
          <el-descriptions-item label="请求">
            {{ detail.requestMethod }} {{ detail.requestPath }}
          </el-descriptions-item>
          <el-descriptions-item label="IP">{{ detail.ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="UA">{{ detail.userAgent || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作时间">{{ detail.createdAt }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">请求参数</el-divider>
        <pre class="detail-json">{{ formatDetail(detail.detail) }}</pre>
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getAuditLogDetail, type AuditLogItem } from '@/api/auditLog'
import { usePermission } from '@/composables/usePermission'
import { PERM, type PermissionCode } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { showApiError } from '@/utils/apiError'
import { findMenuPath } from '@/utils/navigation'

const router = useRouter()
const employeeStore = useEmployeeStore()
const userDrawer = useUserDrawerStore()
const { can } = usePermission()

const visible = ref(false)
const loading = ref(false)
const detail = ref<AuditLogItem | null>(null)

/** 目标类型 → 列表页菜单及定位参数名。 */
const TARGET_PAGES: Record<string, { code: PermissionCode; param: string }> = {
  post: { code: PERM.PostPost, param: 'id' },
  room: { code: PERM.ChatRoom, param: 'id' },
  report: { code: PERM.UserReport, param: 'reviewId' },
  feedback: { code: PERM.UserFeedback, param: 'detailId' },
  employee: { code: PERM.OrgEmployee, param: 'id' },
}

const targetLink = computed(() => {
  const d = detail.value
  if (!d || !d.targetId) return null
  const id = d.targetId
  if (d.targetType === 'user') {
    if (!can(PERM.UserAccountDetail)) return null
    return {
      open: () => {
        visible.value = false
        userDrawer.open(id)
      },
    }
  }
  const page = TARGET_PAGES[d.targetType]
  const path = page && findMenuPath(employeeStore.permissions, page.code)
  if (!page || !path) return null
  return {
    open: () => {
      visible.value = false
      void router.push({ path, query: { [page.param]: String(id) } })
    },
  }
})

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
.audit-detail {
  min-height: 120px;
}

.target-link {
  margin-left: 8px;
}

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
