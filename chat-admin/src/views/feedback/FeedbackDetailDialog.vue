<template>
  <el-dialog v-model="visible" title="反馈详情" width="640px">
    <template v-if="current">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户">
          {{ current.userNickname }} / {{ current.userMobile }}
        </el-descriptions-item>
        <el-descriptions-item label="类型">{{ current.typeText }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ current.statusText }}</el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ current.createdAt }}</el-descriptions-item>
        <el-descriptions-item label="联系方式" :span="2">{{
          current.contact || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">{{ current.content }}</el-descriptions-item>
        <el-descriptions-item label="处理人">{{
          current.employeeName || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ current.handledAt || '-' }}</el-descriptions-item>
        <el-descriptions-item label="处理说明" :span="2">{{
          current.reply || '-'
        }}</el-descriptions-item>
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
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getFeedbackDetail, type FeedbackItem } from '@/api/feedback'
import { showApiError } from '@/utils/apiError'

const visible = ref(false)
const current = ref<FeedbackItem | null>(null)

async function open(row: FeedbackItem) {
  try {
    const res = await getFeedbackDetail(row.id)
    current.value = res.data
    visible.value = true
  } catch (e) {
    showApiError(e, '获取反馈详情失败')
  }
}

defineExpose({ open })
</script>
