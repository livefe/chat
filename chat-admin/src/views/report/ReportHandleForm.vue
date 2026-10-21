<template>
  <section class="review-section handle-section">
    <div class="section-title">处理举报</div>
    <div class="preset-row">
      <el-button size="small" @click="applyPreset('reject')">驳回</el-button>
      <el-button size="small" @click="applyPreset('warn')">警告结案</el-button>
      <el-button
        v-if="report.targetType === ReportTargetType.Post"
        size="small"
        type="warning"
        @click="applyPreset('disablePost')"
      >
        下架并结案
      </el-button>
      <el-button
        v-if="report.targetUserId"
        size="small"
        type="warning"
        @click="applyPreset('mute')"
      >
        禁言 24h 并结案
      </el-button>
      <el-button
        v-if="report.targetType === ReportTargetType.Message"
        size="small"
        type="danger"
        @click="applyPreset('deleteMsg')"
      >
        删除并结案
      </el-button>
      <el-button
        v-if="report.targetUserId"
        size="small"
        type="danger"
        @click="applyPreset('disableUser')"
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
        v-if="form.status === ReportStatus.Resolved && form.action === 3"
        label="禁言时长"
      >
        <el-input-number v-model="form.muteHours" :min="1" :max="8760" />
        <span class="inline-hint">小时</span>
      </el-form-item>
      <el-form-item label="处理说明">
        <el-input
          v-model="form.handleResult"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="填写处理说明（选填）"
        />
      </el-form-item>
    </el-form>
    <div class="handle-footer dialog-footer">
      <el-button @click="emit('cancel')">关闭</el-button>
      <el-button type="primary" :loading="saving" @click="submit">提交处理</el-button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { handleReport, type ReportItem } from '@/api/report'
import { REPORT_HANDLE_ACTION_OPTIONS, ReportStatus, ReportTargetType } from '@/constants/report'
import { showApiError } from '@/utils/apiError'

type HandlePreset = 'reject' | 'warn' | 'mute' | 'disableUser' | 'deleteMsg' | 'disablePost'

const props = defineProps<{ report: ReportItem }>()
const emit = defineEmits<{ handled: []; cancel: [] }>()

const saving = ref(false)
const form = reactive({
  status: ReportStatus.Resolved as number,
  handleResult: '',
  action: 0,
  muteHours: 24,
})

const allowedActions = computed(() => {
  const targetType = props.report.targetType
  return REPORT_HANDLE_ACTION_OPTIONS.filter((opt) => {
    if (opt.value === 2) return targetType === ReportTargetType.Message
    if (opt.value === 5) return targetType === ReportTargetType.Post
    if (opt.value === 3 || opt.value === 4) return true
    return opt.value <= 1
  })
})

function applyPreset(preset: HandlePreset) {
  switch (preset) {
    case 'reject':
      form.status = ReportStatus.Rejected
      form.action = 0
      break
    case 'warn':
      form.status = ReportStatus.Resolved
      form.action = 1
      form.handleResult = form.handleResult || '已警告被举报用户'
      break
    case 'mute':
      form.status = ReportStatus.Resolved
      form.action = 3
      form.muteHours = 24
      form.handleResult = form.handleResult || '禁言 24 小时'
      break
    case 'disableUser':
      form.status = ReportStatus.Resolved
      form.action = 4
      form.handleResult = form.handleResult || '禁用账号'
      break
    case 'deleteMsg':
      form.status = ReportStatus.Resolved
      form.action = 2
      form.handleResult = form.handleResult || '删除违规消息'
      break
    case 'disablePost':
      form.status = ReportStatus.Resolved
      form.action = 5
      form.handleResult = form.handleResult || '下架违规内容'
      break
  }
}

async function submit() {
  saving.value = true
  try {
    await handleReport({
      id: props.report.id,
      status: form.status,
      handleResult: form.handleResult,
      action: form.status === ReportStatus.Resolved ? form.action : 0,
      muteHours: form.muteHours,
    })
    ElMessage.success('处理成功')
    emit('handled')
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
.handle-form {
  margin-top: 4px;
}
.handle-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}
.inline-hint {
  margin-left: 8px;
  color: var(--el-text-color-secondary);
}
</style>
