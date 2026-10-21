<template>
  <div class="toolbar">
    <el-radio-group v-if="!circleOnly" v-model="shape" size="small" @change="emit('shape-change')">
      <el-radio-button :value="1">圆形</el-radio-button>
      <el-radio-button :value="2">多边形</el-radio-button>
    </el-radio-group>

    <template v-if="shape === 1">
      <span class="label">半径：</span>
      <el-input-number
        v-model="radius"
        :min="minRadius"
        :max="maxRadius"
        :step="50"
        size="small"
        @change="emit('radius-change')"
      />
      <span>米</span>
    </template>

    <template v-else>
      <el-button size="small" type="primary" @click="emit('start-polygon')">开始绘制</el-button>
      <el-button size="small" @click="emit('clear-polygon')">清除</el-button>
      <span v-if="polygonHint" class="tip">{{ polygonHint }}</span>
    </template>

    <el-input
      v-model="searchKey"
      placeholder="搜索地点"
      size="small"
      clearable
      style="width: 180px; margin-left: auto"
      @keyup.enter="emit('search', searchKey)"
    />
    <el-button size="small" @click="emit('search', searchKey)">搜索</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  circleOnly: boolean
  minRadius: number
  maxRadius: number
  polygonHint: string
}>()

const shape = defineModel<1 | 2>('shape', { required: true })
const radius = defineModel<number>('radius', { required: true })

const emit = defineEmits<{
  'shape-change': []
  'radius-change': []
  'start-polygon': []
  'clear-polygon': []
  search: [keyword: string]
}>()

const searchKey = ref('')
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.label {
  margin-left: 4px;
  color: var(--el-text-color-regular);
}
.tip {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
