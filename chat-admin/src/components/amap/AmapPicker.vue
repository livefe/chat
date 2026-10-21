<template>
  <div class="amap-picker">
    <AmapPickerToolbar
      v-if="!readonly"
      v-model:shape="shape"
      v-model:radius="circleRadius"
      :circle-only="circleOnly"
      :min-radius="minRadius"
      :max-radius="maxRadius"
      :polygon-hint="polygonHint"
      @shape-change="onShapeChange"
      @radius-change="updateCircle"
      @start-polygon="startPolygon"
      @clear-polygon="clearPolygon"
      @search="doSearch"
    />

    <div v-else class="toolbar-readonly">
      <span class="tip">{{ resolvedHint }}</span>
    </div>

    <div ref="mapRef" class="map">
      <div v-if="mapError" class="map-error">
        <p>{{ mapError }}</p>
        <p class="map-error__hint">
          请在 <code>.env.local</code> 配置 <code>VITE_AMAP_KEY</code> 后重新构建 admin 镜像
        </p>
      </div>
    </div>

    <div class="summary">
      <template v-if="shape === 1 && circleCenter">
        中心：{{ circleCenter[1].toFixed(5) }}, {{ circleCenter[0].toFixed(5) }} · 半径：{{
          circleRadius
        }}m
      </template>
      <template v-else-if="shape === 2 && polygonPoints.length">
        多边形顶点：{{ polygonPoints.length }} 个
      </template>
      <template v-else>
        <span class="tip">在地图上{{ shape === 1 ? '点击选择中心' : '绘制多边形' }}</span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAmapPicker, type AmapPickerChange } from '@/composables/useAmapPicker'
import type { MapRoomOverlay } from '@/utils/mapOverlay'
import AmapPickerToolbar from './AmapPickerToolbar.vue'

interface Props {
  shapeType: 1 | 2
  centerLat?: number
  centerLng?: number
  radius?: number
  polygon?: [number, number][]

  circleOnly?: boolean
  minRadius?: number
  maxRadius?: number

  readonly?: boolean
  readonlyHint?: string

  overlays?: MapRoomOverlay[]
}

const props = withDefaults(defineProps<Props>(), {
  shapeType: 1,
  centerLat: 0,
  centerLng: 0,
  radius: 500,
  polygon: () => [],
  circleOnly: false,
  minRadius: 100,
  maxRadius: 3000,
  readonly: false,
  readonlyHint: '',
  overlays: () => [],
})

const emit = defineEmits<{
  change: [payload: AmapPickerChange]
}>()

const mapRef = ref<HTMLDivElement | null>(null)
const resolvedHint = computed(() => {
  if (props.readonlyHint) return props.readonlyHint
  const shapeLabel = Number(props.shapeType) === 2 ? '蓝色多边形' : '蓝色圆形'
  return `${shapeLabel}为本聊天区，灰色为附近其他聊天区，橙色为疑似重复`
})

const {
  shape,
  circleCenter,
  circleRadius,
  polygonPoints,
  polygonHint,
  mapError,
  updateCircle,
  startPolygon,
  clearPolygon,
  onShapeChange,
  doSearch,
} = useAmapPicker(props, mapRef, (payload) => emit('change', payload))
</script>

<style scoped>
.amap-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.toolbar-readonly {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  min-height: 28px;
}
.tip {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.map {
  width: 100%;
  height: 420px;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-light);
  position: relative;
}
.map-error {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  text-align: center;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  font-size: 14px;
}
.map-error__hint {
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.map-error code {
  font-size: 12px;
  color: var(--el-color-primary);
}
.summary {
  color: var(--el-text-color-regular);
  font-size: 12px;
}
.summary .tip {
  color: var(--el-text-color-secondary);
}
</style>
