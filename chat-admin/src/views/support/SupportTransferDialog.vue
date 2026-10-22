<template>
  <el-dialog v-model="visible" :title="title" width="440px" destroy-on-close>
    <el-select
      v-model="toId"
      filterable
      placeholder="选择坐席"
      style="width: 100%"
      :loading="agentsLoading"
      no-data-text="暂无可转接的坐席"
    >
      <el-option v-for="a in candidates" :key="a.id" :label="`${a.name}（#${a.id}）`" :value="a.id">
        <span class="agent-option">
          <span class="agent-option__dot" :class="{ 'is-online': a.online }" />
          <span>{{ a.name }}（#{{ a.id }}）</span>
          <span class="agent-option__sub">
            {{ a.online ? '在线' : '离线' }}{{ a.departmentName ? ` · ${a.departmentName}` : '' }}
          </span>
        </span>
      </el-option>
    </el-select>
    <div v-if="selected && !selected.online" class="agent-warn">
      {{ selected.name }} 当前不在线，转接后可能无法及时回复用户
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" :disabled="!toId" @click="confirm">
        确认转接
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { transferSupport } from '@/api/support'
import { useSupportAgents } from '@/composables/useSupportOverview'
import { showApiError } from '@/utils/apiError'

const props = defineProps<{
  title: string
  conversationId?: number
  excludeIds: number[]
}>()

const visible = defineModel<boolean>({ required: true })
const emit = defineEmits<{ transferred: [] }>()

const { agents, agentsLoading, loadAgents } = useSupportAgents()
const submitting = ref(false)
const toId = ref<number | undefined>()

const candidates = computed(() => {
  const exclude = new Set(props.excludeIds)
  return agents.value.filter((a) => !exclude.has(a.id))
})
const selected = computed(() => candidates.value.find((a) => a.id === toId.value))

async function confirm() {
  if (!props.conversationId || !toId.value) return
  submitting.value = true
  try {
    await transferSupport(props.conversationId, toId.value)
    visible.value = false
    emit('transferred')
  } catch (e) {
    showApiError(e, '转接失败')
  } finally {
    submitting.value = false
  }
}

watch(visible, (open) => {
  if (!open) return
  toId.value = undefined
  void loadAgents()
})
</script>

<style scoped>
.agent-option {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.agent-option__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-text-color-placeholder);
}

.agent-option__dot.is-online {
  background: var(--el-color-success);
}

.agent-option__sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.agent-warn {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-color-warning);
}
</style>
