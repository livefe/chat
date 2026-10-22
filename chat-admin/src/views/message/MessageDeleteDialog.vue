<template>
  <el-dialog
    v-model="visible"
    :title="messages.length > 1 ? `批量删除消息（${messages.length} 条）` : '删除消息'"
    width="540px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-form label-width="96px">
      <el-form-item label="消息">
        <span v-if="messages.length === 1 && messages[0]">
          #{{ messages[0].id }}（{{ messages[0].fromName }}）
        </span>
        <span v-else>{{ messages.length }} 条，涉及 {{ senderCount }} 位发送者</span>
      </el-form-item>
      <el-form-item label="删除原因" required>
        <div class="reason-field">
          <div class="reason-presets">
            <el-tag
              v-for="r in REASON_PRESETS"
              :key="r"
              class="reason-preset"
              effect="plain"
              @click="reason = r"
            >
              {{ r }}
            </el-tag>
          </div>
          <el-input
            v-model="reason"
            type="textarea"
            :rows="2"
            maxlength="200"
            show-word-limit
            placeholder="请输入删除原因（将记入发送者的处罚记录）"
          />
        </div>
      </el-form-item>
      <el-form-item v-if="canMute" label="禁言发送者">
        <el-switch v-model="muteSender" />
      </el-form-item>
      <el-form-item v-if="canMute && muteSender" label="禁言时长">
        <MuteDurationInput v-model="muteDuration" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="danger" :loading="saving" @click="submit">
          {{ muteSender && canMute ? '删除并禁言' : '确定删除' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { deleteMessages, type AdminMessageItem } from '@/api/message'
import MuteDurationInput from '@/components/user/MuteDurationInput.vue'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'

const REASON_PRESETS = ['垃圾广告', '色情低俗', '辱骂骚扰', '违法违规', '诈骗引流']

const emit = defineEmits<{ success: [] }>()

const { can } = usePermission()
const canMute = computed(() => can(PERM.UserAccountMute))

const visible = ref(false)
const saving = ref(false)
const messages = ref<AdminMessageItem[]>([])
const reason = ref('')
const muteSender = ref(false)
const muteDuration = ref(24)

const senderCount = computed(() => new Set(messages.value.map((m) => m.fromUserId)).size)

function open(rows: AdminMessageItem[]) {
  messages.value = rows.filter((r) => !r.deleted)
  if (messages.value.length === 0) {
    ElMessage.info('所选消息均已删除')
    return
  }
  reason.value = ''
  muteSender.value = false
  muteDuration.value = 24
  visible.value = true
}

async function submit() {
  const text = reason.value.trim()
  if (!text) {
    ElMessage.warning('请填写删除原因')
    return
  }
  saving.value = true
  try {
    const { data } = await deleteMessages({
      ids: messages.value.map((m) => m.id),
      reason: text,
      muteDuration: canMute.value && muteSender.value ? muteDuration.value : 0,
    })
    const parts = [`已删除 ${data.deleted} 条`]
    if (data.muted) parts.push(`禁言 ${data.muted} 人`)
    if (data.skipped) parts.push(`跳过 ${data.skipped} 条`)
    ElMessage.success(parts.join('，'))
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '删除失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.reason-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.reason-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.reason-preset {
  cursor: pointer;
}
</style>
