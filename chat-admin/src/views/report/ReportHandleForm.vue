<template>
  <section class="review-section handle-section">
    <div class="section-title">处理举报</div>
    <div class="preset-row">
      <el-button size="small" @click="applyPreset(ReportStatus.Rejected, ReportHandleAction.None)">
        驳回
      </el-button>
      <el-button
        v-if="report.targetUserId"
        size="small"
        @click="applyPreset(ReportStatus.Resolved, ReportHandleAction.Warn)"
      >
        警告结案
      </el-button>
      <el-button
        v-if="report.targetUserId"
        size="small"
        type="warning"
        @click="applyPreset(ReportStatus.Resolved, ReportHandleAction.Mute)"
      >
        禁言结案
      </el-button>
      <el-button
        v-if="report.targetType === ReportTargetType.Message"
        size="small"
        type="danger"
        @click="applyPreset(ReportStatus.Resolved, ReportHandleAction.DeleteContent)"
      >
        删除并结案
      </el-button>
      <el-button
        v-if="report.targetType === ReportTargetType.Post"
        size="small"
        type="warning"
        @click="applyPreset(ReportStatus.Resolved, ReportHandleAction.DisablePost)"
      >
        下架并结案
      </el-button>
      <el-button
        v-if="report.targetUserId"
        size="small"
        type="danger"
        @click="applyPreset(ReportStatus.Resolved, ReportHandleAction.DisableUser)"
      >
        禁用账号并结案
      </el-button>
    </div>
    <el-form :model="form" label-width="88px" class="handle-form">
      <el-form-item label="处理结果">
        <el-radio-group v-model="form.status">
          <el-radio :value="ReportStatus.Resolved">已处理</el-radio>
          <el-radio :value="ReportStatus.Rejected">驳回</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.status === ReportStatus.Resolved" label="处置动作">
        <el-select v-model="form.action" style="width: 100%">
          <el-option
            v-for="opt in allowedActions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        v-if="form.status === ReportStatus.Resolved && form.action === ReportHandleAction.Mute"
        label="禁言时长"
      >
        <MuteDurationInput v-model="form.muteDuration" />
      </el-form-item>
      <el-form-item label="处理说明">
        <el-input
          v-model="form.handleResult"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          :placeholder="handleResultPlaceholder"
        />
      </el-form-item>
      <el-form-item v-if="(report.targetReportCount ?? 0) > 1" label="同对象">
        <el-checkbox v-model="form.resolveSameTarget">
          其余 {{ (report.targetReportCount ?? 1) - 1 }} 条举报中待处理的按相同结果一并结案
        </el-checkbox>
      </el-form-item>
    </el-form>
    <div class="handle-footer">
      <el-checkbox v-model="openNext">处理后打开下一条待处理</el-checkbox>
      <div class="dialog-footer">
        <el-button @click="emit('cancel')">关闭</el-button>
        <el-button type="primary" :loading="saving" @click="submit">提交处理</el-button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { handleReport, type ReportItem } from '@/api/report'
import MuteDurationInput from '@/components/user/MuteDurationInput.vue'
import {
  REPORT_HANDLE_ACTION_OPTIONS,
  ReportHandleAction,
  ReportStatus,
  ReportTargetType,
} from '@/constants/report'
import { muteDurationLabel } from '@/constants/user'
import { showApiError } from '@/utils/apiError'

const OPEN_NEXT_KEY = 'report:openNext'

const props = defineProps<{ report: ReportItem }>()
const emit = defineEmits<{ handled: [openNext: boolean]; cancel: [] }>()

const saving = ref(false)
const openNext = ref(localStorage.getItem(OPEN_NEXT_KEY) !== '0')
const form = reactive({
  status: ReportStatus.Resolved as number,
  action: ReportHandleAction.None as number,
  muteDuration: 24,
  handleResult: '',
  resolveSameTarget: false,
})

const allowedActions = computed(() =>
  REPORT_HANDLE_ACTION_OPTIONS.filter((opt) => {
    switch (opt.value) {
      case ReportHandleAction.DeleteContent:
        return props.report.targetType === ReportTargetType.Message
      case ReportHandleAction.DisablePost:
        return props.report.targetType === ReportTargetType.Post
      case ReportHandleAction.Warn:
      case ReportHandleAction.Mute:
      case ReportHandleAction.DisableUser:
        return props.report.targetUserId > 0
      default:
        return true
    }
  }),
)

const handleResultPlaceholder = computed(() =>
  form.action === ReportHandleAction.Warn
    ? '警告原因会以站内消息发送给用户；留空时按举报原因生成'
    : '选填；会记入被举报用户的处罚记录',
)

const actionLabel = (action: number) =>
  REPORT_HANDLE_ACTION_OPTIONS.find((o) => o.value === action)?.label ?? ''

function applyPreset(status: number, action: number) {
  form.status = status
  form.action = action
  if (action === ReportHandleAction.Mute) form.muteDuration = 24
}

async function submit() {
  const resolved = form.status === ReportStatus.Resolved
  const action = resolved ? form.action : ReportHandleAction.None
  const parts = [resolved ? '确认处理该举报' : '确认驳回该举报']
  if (action === ReportHandleAction.Mute) {
    parts.push(`禁言被举报用户 ${muteDurationLabel(form.muteDuration)}`)
  } else if (action !== ReportHandleAction.None) {
    parts.push(actionLabel(action))
  }
  if (form.resolveSameTarget) parts.push('同对象的其它待处理举报一并结案')
  try {
    await ElMessageBox.confirm(`${parts.join('，')}？`, '提交处理', {
      type: action === ReportHandleAction.None ? 'info' : 'warning',
    })
  } catch {
    return
  }

  saving.value = true
  try {
    const { data } = await handleReport({
      id: props.report.id,
      status: form.status,
      handleResult: form.handleResult.trim(),
      action,
      muteDuration: action === ReportHandleAction.Mute ? form.muteDuration : 0,
      resolveSameTarget: form.resolveSameTarget,
    })
    ElMessage.success(
      data.resolvedCount > 0
        ? `处理成功，另有 ${data.resolvedCount} 条同对象举报已结案`
        : '处理成功',
    )
    localStorage.setItem(OPEN_NEXT_KEY, openNext.value ? '1' : '0')
    emit('handled', openNext.value)
  } catch (e) {
    showApiError(e, '处理失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.handle-section {
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 14px;
}

.preset-row .el-button + .el-button {
  margin-left: 0;
}

.handle-form {
  margin-top: 4px;
}

.handle-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 8px;
}
</style>
