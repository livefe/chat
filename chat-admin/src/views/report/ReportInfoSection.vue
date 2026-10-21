<template>
  <section class="review-section">
    <div class="section-title">举报信息</div>
    <el-descriptions :column="2" border size="small">
      <el-descriptions-item label="举报人">
        {{ report.reporterNickname }} / {{ report.reporterMobile }}
      </el-descriptions-item>
      <el-descriptions-item label="被举报用户">
        {{ report.targetUserName || '-' }}
      </el-descriptions-item>
      <el-descriptions-item label="对象">
        {{ report.targetTypeText }} #{{ report.targetId }}
      </el-descriptions-item>
      <el-descriptions-item label="原因">{{ report.reasonText }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="reportStatusTagType(report.status)" size="small">{{
          report.statusText
        }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="提交时间">{{ report.createdAt }}</el-descriptions-item>
      <el-descriptions-item v-if="report.content" label="补充说明" :span="2">
        {{ report.content }}
      </el-descriptions-item>
    </el-descriptions>
    <div v-if="report.images?.length" class="images">
      <div class="images-label">举证图片</div>
      <el-image
        v-for="(url, i) in report.images"
        :key="i"
        :src="url"
        :preview-src-list="report.images"
        fit="cover"
        class="thumb"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import type { ReportItem } from '@/api/report'
import { reportStatusTagType } from '@/utils/reportDisplay'

defineProps<{ report: ReportItem }>()
</script>
