<template>
  <el-dialog v-model="visible" :title="`结案「${sessionTitle}」`" width="440px" destroy-on-close>
    <el-form label-width="72px" @submit.prevent>
      <el-form-item label="结案原因">
        <el-radio-group v-model="reason">
          <el-radio-button
            v-for="opt in SUPPORT_CLOSE_REASON_OPTIONS"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="备注">
        <el-input
          v-model="note"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          placeholder="选填，仅内部可见"
        />
      </el-form-item>
    </el-form>
    <div class="close-hint">结案后用户需重新发起咨询，本会话只能查看记录。</div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="danger" :loading="submitting" @click="confirm">确认结案</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { closeSupport } from '@/api/support'
import { SUPPORT_CLOSE_REASON_OPTIONS } from '@/constants/support'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{ conversationId?: number; sessionTitle: string }>()
const visible = defineModel<boolean>({ required: true })
const emit = defineEmits<{ closed: [reason: string, note: string] }>()

const reason = ref('resolved')
const note = ref('')
const submitting = ref(false)

watch(visible, (open) => {
  if (!open) return
  reason.value = 'resolved'
  note.value = ''
})

async function confirm() {
  if (!props.conversationId) return
  submitting.value = true
  const trimmedNote = note.value.trim()
  try {
    await closeSupport({
      conversationId: props.conversationId,
      reason: reason.value,
      note: trimmedNote,
    })
    visible.value = false
    emit('closed', reason.value, trimmedNote)
  } catch (e) {
    showApiError(e, '结案失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.close-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
