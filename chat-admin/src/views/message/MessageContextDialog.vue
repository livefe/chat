<template>
  <el-dialog v-model="visible" title="消息上下文" width="720px">
    <div v-loading="loading" class="context-list">
      <div
        v-for="item in list"
        :key="item.id"
        class="context-item"
        :class="{ highlight: item.id === centerId }"
      >
        <div class="context-meta">
          <span class="name">{{ item.fromName }}</span>
          <span class="time">{{ formatUnixSecondsLocale(item.createdAt) }}</span>
          <el-tag size="small" effect="plain">{{ item.scopeText }}</el-tag>
          <el-tag v-if="item.deleted" type="danger" size="small">已删除</el-tag>
          <el-tag v-if="item.recalled" type="warning" size="small">已撤回</el-tag>
        </div>
        <AdminReplyQuote :item="item" />
        <div class="context-content" :class="{ 'muted-content': item.deleted || item.recalled }">
          <MessageContent :item="item" />
        </div>
      </div>
      <el-empty v-if="!loading && list.length === 0" description="暂无上下文" />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getMessageContext, type AdminMessageItem } from '@/api/message'
import AdminReplyQuote from '@/components/AdminReplyQuote.vue'
import { showApiError } from '@/utils/apiError'
import { formatUnixSecondsLocale } from '@/utils/datetime'
import MessageContent from './MessageContent.vue'

const visible = ref(false)
const loading = ref(false)
const list = ref<AdminMessageItem[]>([])
const centerId = ref(0)

async function open(row: AdminMessageItem) {
  centerId.value = row.id
  visible.value = true
  loading.value = true
  try {
    const res = await getMessageContext(row.id)
    list.value = res.data.list ?? []
  } catch (e) {
    showApiError(e, '获取上下文失败')
    list.value = []
  } finally {
    loading.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.context-list {
  max-height: 480px;
  overflow-y: auto;
}
.context-item {
  padding: 10px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.context-item.highlight {
  background: var(--el-color-primary-light-9);
}
.context-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}
.context-meta .name {
  font-weight: 600;
}
.context-meta .time {
  color: var(--el-text-color-secondary);
}
.context-content {
  color: var(--el-text-color-primary);
  line-height: 1.5;
  word-break: break-word;
}
.muted-content {
  color: var(--el-text-color-secondary);
  text-decoration: line-through;
}
</style>
