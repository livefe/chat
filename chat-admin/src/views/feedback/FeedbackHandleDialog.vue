<template>
  <el-dialog v-model="visible" title="处理反馈" width="480px">
    <el-form :model="form" label-width="80px">
      <el-form-item label="状态">
        <el-select v-model="form.status" style="width: 100%">
          <el-option
            v-for="opt in allowedStatusOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="处理说明">
        <el-input v-model="form.reply" type="textarea" :rows="4" maxlength="500" show-word-limit />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { updateFeedbackStatus, type FeedbackItem } from '@/api/feedback'
import { FEEDBACK_STATUS_OPTIONS, FeedbackStatus } from '@/constants/feedback'
import { showApiError } from '@/utils/apiError'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const saving = ref(false)
const form = reactive({ id: 0, status: FeedbackStatus.Resolved as number, reply: '' })

const allowedStatusOptions = FEEDBACK_STATUS_OPTIONS.filter(
  (o) => o.value !== FeedbackStatus.Pending,
)

function open(row: FeedbackItem) {
  form.id = row.id
  form.status = row.status === FeedbackStatus.Pending ? FeedbackStatus.Resolved : row.status
  form.reply = row.reply || ''
  visible.value = true
}

async function submit() {
  saving.value = true
  try {
    await updateFeedbackStatus({ id: form.id, status: form.status, reply: form.reply })
    ElMessage.success('更新成功')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '更新失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>
