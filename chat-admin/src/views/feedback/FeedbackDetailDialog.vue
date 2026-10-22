<template>
  <el-dialog v-model="visible" title="反馈详情" width="660px" destroy-on-close>
    <div v-loading="loading" class="feedback-detail">
      <template v-if="current">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="用户">
            <UserCell
              :user-id="current.userId"
              :nickname="current.userNickname"
              :mobile="current.userMobile"
            />
          </el-descriptions-item>
          <el-descriptions-item label="类型">{{ current.typeText }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="feedbackStatusTagType(current.status)" size="small">
              {{ current.statusText }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ current.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="联系方式" :span="2">
            {{ current.contact || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="内容" :span="2">{{ current.content }}</el-descriptions-item>
          <template v-if="!pending">
            <el-descriptions-item label="处理人">
              {{ current.employeeName || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理时间">
              {{ current.handledAt || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="处理说明" :span="2">
              {{ current.handleResult || '-' }}
            </el-descriptions-item>
          </template>
        </el-descriptions>
        <div v-if="current.images?.length" class="images">
          <el-image
            v-for="(url, i) in current.images"
            :key="i"
            :src="url"
            :preview-src-list="current.images"
            fit="cover"
            class="thumb"
          />
        </div>

        <el-form v-if="canHandle" label-width="80px" class="handle-form">
          <el-form-item label="处理结果">
            <el-radio-group v-model="form.status">
              <el-radio :value="FeedbackStatus.Resolved">已解决</el-radio>
              <el-radio :value="FeedbackStatus.Closed">关闭</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="处理说明">
            <el-input
              v-model="form.handleResult"
              type="textarea"
              :rows="3"
              maxlength="500"
              show-word-limit
              placeholder="用户可在 App「我的反馈」中看到"
            />
          </el-form-item>
        </el-form>
      </template>
    </div>
    <template v-if="canHandle" #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">提交处理</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { getFeedbackDetail, updateFeedbackStatus, type FeedbackItem } from '@/api/feedback'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { FeedbackStatus, feedbackStatusTagType } from '@/constants/feedback'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'

const emit = defineEmits<{ success: [] }>()

const { can } = usePermission()

const visible = ref(false)
const loading = ref(false)
const saving = ref(false)
const current = ref<FeedbackItem | null>(null)
const form = reactive({ status: FeedbackStatus.Resolved as number, handleResult: '' })

const pending = computed(() => current.value?.status === FeedbackStatus.Pending)
const canHandle = computed(() => pending.value && can(PERM.UserFeedbackStatus))

async function open(id: number) {
  visible.value = true
  loading.value = true
  current.value = null
  form.status = FeedbackStatus.Resolved
  form.handleResult = ''
  try {
    const res = await getFeedbackDetail(id)
    current.value = res.data
  } catch (e) {
    visible.value = false
    showApiError(e, '获取反馈详情失败')
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!current.value) return
  saving.value = true
  try {
    await updateFeedbackStatus({
      id: current.value.id,
      status: form.status,
      handleResult: form.handleResult.trim(),
    })
    ElMessage.success('已处理')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '处理失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.feedback-detail {
  min-height: 160px;
}

.handle-form {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
