<template>
  <span class="copy-text">
    <span>{{ text }}</span>
    <el-icon class="copy-text__btn" title="复制" @click.stop="copy"><DocumentCopy /></el-icon>
  </span>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { DocumentCopy } from '@element-plus/icons-vue'

const props = defineProps<{ text: string | number }>()

async function copy() {
  try {
    await navigator.clipboard.writeText(String(props.text))
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败，请手动选择复制')
  }
}
</script>

<style scoped>
.copy-text {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.copy-text__btn {
  cursor: pointer;
  color: var(--el-text-color-placeholder);
}

.copy-text__btn:hover {
  color: var(--el-color-primary);
}
</style>
