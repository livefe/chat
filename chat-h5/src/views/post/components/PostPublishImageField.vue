<script setup lang="ts">
import type { UploaderFileListItem } from 'vant'

defineProps<{
  afterRead: (item: UploaderFileListItem | UploaderFileListItem[]) => void | Promise<void>
}>()

const images = defineModel<UploaderFileListItem[]>({ required: true })
</script>

<template>
  <div class="form-card">
    <van-field class="field--media" label="图片">
      <template #input>
        <van-uploader v-model="images" :after-read="afterRead" :max-count="9" multiple />
      </template>
    </van-field>
  </div>
</template>

<style scoped>
.form-card :deep(.field--media.van-cell) {
  align-items: flex-start;
}
.form-card :deep(.field--media .van-field__label) {
  width: 72px;
  margin-right: 12px;
  padding-top: 2px;
}
.form-card :deep(.field--media .van-field__control--custom) {
  display: block;
  overflow: visible;
}
.form-card :deep(.field--media .van-uploader) {
  display: block;
  width: 100%;
  margin-top: 0;
}
.form-card :deep(.field--media .van-uploader__wrapper) {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.form-card :deep(.field--media .van-uploader__preview),
.form-card :deep(.field--media .van-uploader__upload) {
  flex: 0 0 calc((100% - 16px) / 3);
  width: calc((100% - 16px) / 3);
  height: auto;
  aspect-ratio: 1 / 1;
  margin: 0;
}
.form-card :deep(.field--media .van-uploader__preview-image),
.form-card :deep(.field--media .van-uploader__file),
.form-card :deep(.field--media .van-uploader__upload) {
  width: 100%;
  height: 100%;
}
.form-card :deep(.field--media .van-uploader__preview-delete) {
  transform: scale(0.9);
  transform-origin: top right;
}
</style>
