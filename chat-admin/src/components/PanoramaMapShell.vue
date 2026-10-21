<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="92vw"
    top="4vh"
    destroy-on-close
    class="panorama-map-dialog"
  >
    <div
      v-loading="loading"
      class="map-body"
      :style="{ gridTemplateColumns: `1fr ${sideWidth}px` }"
    >
      <div class="map-toolbar">
        <slot name="toolbar" />
      </div>

      <div ref="canvasRef" class="map-canvas">
        <div v-if="error" class="map-error">
          <p>{{ error }}</p>
          <p class="map-error__hint">
            请在 <code>.env.local</code> 配置 <code>VITE_AMAP_KEY</code> 后重新构建 admin 镜像
          </p>
        </div>
      </div>

      <div class="map-side">
        <slot name="side" />
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  title: string
  loading: boolean
  error: string
  sideWidth: number
}>()

const visible = defineModel<boolean>('visible', { required: true })
const canvasRef = ref<HTMLDivElement | null>(null)

defineExpose({ canvas: canvasRef })
</script>

<style scoped>
.map-body {
  display: grid;
  grid-template-rows: auto 1fr;
  gap: 12px;
  min-height: 72vh;
}

.map-toolbar {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.map-toolbar :slotted(.map-toolbar__meta) {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.map-toolbar :slotted(.map-legend) {
  margin-left: auto;
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.map-toolbar :slotted(.dot) {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
}

.map-canvas {
  position: relative;
  min-height: 64vh;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--el-border-color);
}

.map-error {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  z-index: 2;
}

.map-error__hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.map-side {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 12px;
  background: var(--el-bg-color);
  align-self: stretch;
}

.map-side :slotted(.map-side__title) {
  font-weight: 600;
  margin-bottom: 12px;
  word-break: break-all;
}

.map-side :slotted(.map-side__row) {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 13px;
  margin-bottom: 8px;
  color: var(--el-text-color-regular);
}

.map-side :slotted(.map-side__row span:first-child) {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.map-side :slotted(.map-side__empty) {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.6;
}
</style>

<style>
.panorama-map-dialog .el-dialog__body {
  padding-top: 8px;
}
</style>
