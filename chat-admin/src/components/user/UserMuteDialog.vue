<template>
  <el-dialog
    v-model="muteDialog.visible"
    :title="isBatch ? `批量禁言（${muteDialog.users.length} 人）` : '禁言用户'"
    width="520px"
    destroy-on-close
    :close-on-click-modal="false"
    @open="reset"
  >
    <el-form label-width="80px">
      <el-form-item label="对象">
        <span v-if="!isBatch && first">{{ userLabel(first) }}</span>
        <span v-else class="sub">{{ batchLabel }}</span>
      </el-form-item>
      <el-form-item label="时长" required>
        <MuteDurationInput v-model="duration" />
      </el-form-item>
      <el-form-item label="原因" required>
        <el-input
          v-model="reason"
          type="textarea"
          :rows="3"
          maxlength="200"
          show-word-limit
          placeholder="请输入禁言原因（将记入处罚记录）"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="muteDialog.visible = false">取消</el-button>
        <el-button type="warning" :loading="saving" @click="submit">确定禁言</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'

import { batchMuteUser, muteUser } from '@/api/user'
import { batchResultText, useUserActionDialogs, userLabel } from '@/composables/useUserActions'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { showApiError } from '@/utils/apiError'
import MuteDurationInput from './MuteDurationInput.vue'

const { muteDialog } = useUserActionDialogs()
const drawer = useUserDrawerStore()

const duration = ref(24)
const reason = ref('')
const saving = ref(false)

const isBatch = computed(() => muteDialog.users.length > 1)
const first = computed(() => muteDialog.users[0])
const batchLabel = computed(() => {
  const names = muteDialog.users.slice(0, 5).map((u) => u.nickname || `ID ${u.id}`)
  const more = muteDialog.users.length > 5 ? ` 等 ${muteDialog.users.length} 人` : ''
  return names.join('、') + more
})

function reset() {
  duration.value = 24
  reason.value = ''
}

async function submit() {
  const text = reason.value.trim()
  if (!text) {
    ElMessage.warning('请填写禁言原因')
    return
  }
  saving.value = true
  try {
    if (isBatch.value) {
      const { data } = await batchMuteUser(
        muteDialog.users.map((u) => u.id),
        duration.value,
        text,
      )
      ElMessage.success(batchResultText('禁言', data))
    } else if (first.value) {
      await muteUser(first.value.id, duration.value, text)
      ElMessage.success('已禁言')
    }
    muteDialog.visible = false
    drawer.refresh()
  } catch (e) {
    showApiError(e, '禁言失败')
  } finally {
    saving.value = false
  }
}
</script>
