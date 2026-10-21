<template>
  <el-dialog v-model="visible" :title="title" width="420px" destroy-on-close>
    <el-select
      v-model="toId"
      filterable
      placeholder="选择同事"
      style="width: 100%"
      :loading="loading"
    >
      <el-option
        v-for="c in contacts"
        :key="c.id"
        :label="`${c.name}（#${c.id}）${c.departmentName ? ' · ' + c.departmentName : ''}`"
        :value="c.id"
      />
    </el-select>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="confirm">确认转接</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getImContacts, type ImContactItem } from '@/api/im'
import { transferSupport } from '@/api/support'
import { isCancelAction } from '@/utils/confirm'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{
  title: string
  conversationId?: number
  excludeIds: number[]
}>()

const visible = defineModel<boolean>({ required: true })
const emit = defineEmits<{ transferred: [] }>()

const loading = ref(false)
const submitting = ref(false)
const contacts = ref<ImContactItem[]>([])
const toId = ref<number | undefined>()

async function loadContacts() {
  toId.value = undefined
  loading.value = true
  try {
    const { data } = await getImContacts({ page: 1, pageSize: 200 })
    const exclude = new Set(props.excludeIds)
    contacts.value = (data.list || []).filter((c) => !exclude.has(c.id))
    if (contacts.value.length === 0) {
      ElMessage.warning('暂无可转接同事')
      visible.value = false
    }
  } catch (e) {
    showApiError(e, '加载同事失败')
    visible.value = false
  } finally {
    loading.value = false
  }
}

async function confirm() {
  if (!props.conversationId || !toId.value) {
    ElMessage.warning('请选择转接对象')
    return
  }
  submitting.value = true
  try {
    await ElMessageBox.confirm('确认转接该会话？', '转接')
    await transferSupport(props.conversationId, toId.value)
    visible.value = false
    emit('transferred')
  } catch (e) {
    if (isCancelAction(e)) return
    showApiError(e, '转接失败')
  } finally {
    submitting.value = false
  }
}

watch(visible, (open) => {
  if (open) void loadContacts()
})
</script>
