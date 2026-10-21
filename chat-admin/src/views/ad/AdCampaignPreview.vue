<template>
  <div class="preview-panel">
    <div class="preview-title">预览</div>
    <div class="ad-preview">
      <img v-if="imageUrl" :src="imageUrl" class="preview-image" alt="" />
      <div v-else class="preview-image placeholder">封面</div>
      <div class="preview-text">
        <div class="preview-ad-title">{{ title || '广告标题' }}</div>
        <div class="preview-ad-meta">
          <span class="preview-ad-name">{{ advertiserName || '广告主' }}</span>
          <span class="preview-ad-badge" title="信息流须显著标明广告">广告</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  imageUrl: string
  title: string
  advertiserName: string
}>()
</script>

<style scoped>
.preview-panel {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 12px;
  background: var(--el-fill-color-blank);
}
.preview-title {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 10px;
}
.ad-preview {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.preview-image {
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
  border-radius: 6px;
  background: var(--el-fill-color);
}
.preview-image.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}
.preview-ad-title {
  font-weight: 600;
  font-size: 14px;
  line-height: 1.4;
}
.preview-ad-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 6px;
  min-width: 0;
}
.preview-ad-name {
  flex: 1;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
.preview-ad-badge {
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.06em;
  line-height: 1;
  color: #909090;
  border: 1px solid #d0d0d0;
  border-radius: 2px;
  padding: 2px 4px;
}
</style>
