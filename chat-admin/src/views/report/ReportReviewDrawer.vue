<template>
  <el-drawer v-model="visible" :title="reviewTitle" size="720px" destroy-on-close>
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

        <ReportTargetUserCard
          v-if="current.targetUser"
          :user="current.targetUser"
          @show-reports="showUserReports(current.targetUserId)"
        />

        <ReportInfoSection :report="current" />

        <ReportTargetSection
          :report="current"
          :message-context="messageContext"
          @leave="visible = false"
        />

        <section v-if="current.status !== ReportStatus.Pending" class="review-section">
          <div class="section-title">处理记录</div>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="结果">{{ current.statusText }}</el-descriptions-item>
            <el-descriptions-item label="处置动作">
              {{ current.handleActionText || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理人">
              {{ current.employeeName || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理时间">
              {{ current.handledAt || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理说明" :span="2">
              {{ current.handleResult || '-' }}
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
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getMessageContext, type AdminMessageItem } from '@/api/message'
import { getReportDetail, type ReportItem } from '@/api/report'
import { usePermission } from '@/composables/usePermission'
import { FILTER_ALL } from '@/constants'
import { PERM } from '@/constants/permission'
import { ReportStatus, ReportTargetType } from '@/constants/report'
import { useEmployeeStore } from '@/stores/employee'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { showApiError } from '@/utils/apiError'
import { pushMenuRoute } from '@/utils/navigation'
import ReportHandleForm from './ReportHandleForm.vue'
import ReportInfoSection from './ReportInfoSection.vue'
import ReportTargetSection from './ReportTargetSection.vue'
import ReportTargetUserCard from './ReportTargetUserCard.vue'

const emit = defineEmits<{ handled: [openNext: boolean] }>()

const { can } = usePermission()
const router = useRouter()
const employeeStore = useEmployeeStore()
const drawer = useUserDrawerStore()

const visible = ref(false)
const reviewLoading = ref(false)
const current = ref<ReportItem | null>(null)
const messageContext = ref<AdminMessageItem[]>([])

const reviewTitle = computed(() => (current.value ? `举报审核 #${current.value.id}` : '举报审核'))

const canHandle = computed(
  () =>
    !!current.value && can(PERM.UserReportHandle) && current.value.status === ReportStatus.Pending,
)

async function loadMessageContext(messageId: number) {
  messageContext.value = []
  if (!can(PERM.ChatMessageContext)) return
  try {
    const res = await getMessageContext(messageId, { limit: 6 })
    messageContext.value = res.data.list ?? []
  } catch {
    messageContext.value = []
  }
}

async function load(id: number) {
  reviewLoading.value = true
  try {
    const res = await getReportDetail(id)
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

async function open(id: number) {
  visible.value = true
  current.value = null
  messageContext.value = []
  await load(id)
}

function showUserReports(userId: number) {
  visible.value = false
  pushMenuRoute(router, employeeStore.permissions, PERM.UserReport, {
    targetUserId: String(userId),
    status: String(FILTER_ALL),
  })
}

function onHandled(openNext: boolean) {
  visible.value = false
  emit('handled', openNext)
}

watch(
  () => drawer.changed,
  () => {
    if (visible.value && current.value) void load(current.value.id)
  },
)

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
