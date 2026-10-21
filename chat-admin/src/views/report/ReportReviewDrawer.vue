<template>
  <el-drawer
    v-model="visible"
    :title="reviewTitle"
    size="680px"
    destroy-on-close
    class="review-drawer"
  >
    <div v-loading="reviewLoading" class="review-body">
      <template v-if="current">
        <el-alert
          v-if="(current.targetReportCount ?? 0) > 1"
          :title="`该对象已被举报 ${current.targetReportCount} 次`"
          type="warning"
          :closable="false"
          show-icon
          class="alert-block"
        />

        <ReportInfoSection :report="current" />

        <ReportTargetSection
          :report="current"
          :message-context="messageContext"
          @leave="visible = false"
        />

        <section v-if="current.handleResult || current.handledAt" class="review-section">
          <div class="section-title">处理记录</div>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="处置动作">
              {{ current.handleActionText || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理说明">
              {{ current.handleResult || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理人">
              {{ current.employeeName || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理时间">
              {{ current.handledAt || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </section>

        <ReportHandleForm
          v-if="canHandle"
          :key="current.id"
          :report="current"
          @cancel="visible = false"
          @handled="onHandled"
        />
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { getMessageContext, type AdminMessageItem } from '@/api/message'
import { getReportDetail, type ReportItem } from '@/api/report'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { ReportStatus, ReportTargetType } from '@/constants/report'
import { showApiError } from '@/utils/apiError'
import ReportHandleForm from './ReportHandleForm.vue'
import ReportInfoSection from './ReportInfoSection.vue'
import ReportTargetSection from './ReportTargetSection.vue'

const emit = defineEmits<{ handled: [] }>()

const { can } = usePermission()

const visible = ref(false)
const reviewLoading = ref(false)
const current = ref<ReportItem | null>(null)
const messageContext = ref<AdminMessageItem[]>([])

const reviewTitle = computed(() => {
  if (!current.value) return '举报审核'
  return `举报审核 #${current.value.id}`
})

const canHandle = computed(() => {
  if (!current.value) return false
  if (!can(PERM.UserReportHandle)) return false
  const status = current.value.status
  return status !== ReportStatus.Resolved && status !== ReportStatus.Rejected
})

async function loadMessageContext(messageId: number) {
  messageContext.value = []
  if (!can(PERM.ChatMessageContext)) return
  try {
    const res = await getMessageContext(messageId, 12)
    messageContext.value = res.data?.list ?? []
  } catch {
    messageContext.value = []
  }
}

async function open(row: Pick<ReportItem, 'id'>) {
  visible.value = true
  reviewLoading.value = true
  current.value = null
  messageContext.value = []
  try {
    const res = await getReportDetail(row.id)
    current.value = res.data
    if (res.data.targetType === ReportTargetType.Message) {
      await loadMessageContext(res.data.targetId)
    }
  } catch (e) {
    visible.value = false
    showApiError(e, '获取举报详情失败')
  } finally {
    reviewLoading.value = false
  }
}

function onHandled() {
  visible.value = false
  emit('handled')
}

defineExpose({ open })
</script>

<style scoped>
.alert-block {
  margin-bottom: 12px;
}
.review-body {
  min-height: 200px;
  padding-bottom: 24px;
}
.review-section + .review-section {
  margin-top: 20px;
}
</style>
