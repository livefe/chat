<template>
  <PanoramaMapShell
    ref="shellRef"
    v-model:visible="visible"
    title="聊天区全景地图"
    :loading="loading"
    :error="mapError"
    :side-width="220"
  >
    <template #toolbar>
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 120px">
        <el-option label="待审核" :value="0" />
        <el-option label="正常" :value="1" />
        <el-option label="驳回" :value="2" />
        <el-option label="禁用" :value="3" />
      </el-select>
      <el-select v-model="filterPlaceType" placeholder="类型" clearable style="width: 130px">
        <el-option
          v-for="item in ROOM_PLACE_TYPE_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
      <el-button type="primary" @click="reload">刷新</el-button>
      <el-button @click="fitAll">适应全部</el-button>
      <span class="map-toolbar__meta">
        共 {{ rooms.length }} 个聊天区 · 缩小到最大视野后只能拖动
      </span>
      <div class="map-legend">
        <span><i class="dot dot--ok" />正常</span>
        <span><i class="dot dot--pending" />待审核</span>
        <span><i class="dot dot--reject" />驳回</span>
        <span><i class="dot dot--disabled" />禁用</span>
      </div>
    </template>

    <template #side>
      <template v-if="selected">
        <div class="map-side__title">{{ selected.name }}</div>
        <div class="map-side__row"><span>ID</span>{{ selected.id }}</div>
        <div class="map-side__row"><span>类型</span>{{ selected.placeTypeText || '-' }}</div>
        <div class="map-side__row">
          <span>形状</span>{{ selected.shapeType === 2 ? '多边形' : '圆形' }}
        </div>
        <div class="map-side__row"><span>状态</span>{{ selected.statusText }}</div>
        <div class="map-side__row"><span>来源</span>{{ selected.sourceText || '-' }}</div>
        <div v-if="selected.shapeType !== 2" class="map-side__row">
          <span>半径</span>{{ selected.radius }} 米
        </div>
      </template>
      <div v-else class="map-side__empty">点击地图上的聊天区查看详情</div>
    </template>
  </PanoramaMapShell>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getRoomMap, type RoomMapItem } from '@/api/room'
import PanoramaMapShell from '@/components/PanoramaMapShell.vue'
import { usePanoramaMap } from '@/composables/usePanoramaMap'
import { showApiError } from '@/utils/apiError'
import { hasCoordinates, wgs84ToGcj02 } from '@/utils/geo'
import {
  escapeHtml,
  mapInfoHtml,
  mapLabelHtml,
  pathCenter,
  themeColor,
  wgsPolygonToGcjPath,
  type LngLatTuple,
} from '@/utils/mapOverlay'
import { ROOM_PLACE_TYPE_OPTIONS } from '@/constants/room'

const visible = defineModel<boolean>('visible', { required: true })

const shellRef = ref<InstanceType<typeof PanoramaMapShell>>()
const loading = ref(false)
const rooms = ref<RoomMapItem[]>([])
const selected = ref<RoomMapItem | null>(null)
const filterStatus = ref<number | undefined>(undefined)
const filterPlaceType = ref<number | undefined>(undefined)

const { mapError, ensureMap, paint, openInfo, fitAll } = usePanoramaMap({
  visible: () => visible.value,
  container: () => shellRef.value?.canvas,
  onShow: reload,
  onHide: () => {
    rooms.value = []
    selected.value = null
  },
})

function statusColor(status: number) {
  if (status === 1) return themeColor('--el-color-success', '#67c23a')
  if (status === 0) return themeColor('--el-color-warning', '#e6a23c')
  if (status === 2) return themeColor('--el-text-color-secondary', '#909399')
  return themeColor('--el-color-danger', '#f56c6c')
}

function bindSelect(shape: AMap.Overlay, item: RoomMapItem, center: LngLatTuple) {
  shape.on('click', () => {
    selected.value = item
    openInfo(
      mapInfoHtml(
        escapeHtml(item.name),
        `#${item.id} · ${escapeHtml(item.statusText)} · ${escapeHtml(item.placeTypeText || '')}`,
      ),
      center,
    )
  })
}

function paintRooms() {
  paint((amap, add) => {
    for (const item of rooms.value) {
      const color = statusColor(item.status)
      let shape: AMap.Overlay
      let center: LngLatTuple
      if (item.shapeType === 2 && item.polygon && item.polygon.length >= 3) {
        const path = wgsPolygonToGcjPath(item.polygon)
        center = pathCenter(path)
        shape = new amap.Polygon({
          path,
          strokeColor: color,
          strokeWeight: 2,
          fillColor: color,
          fillOpacity: 0.18,
          zIndex: 50,
          cursor: 'pointer',
        })
      } else if (hasCoordinates(item.centerLat, item.centerLng)) {
        const c = wgs84ToGcj02(item.centerLat, item.centerLng)
        center = [c.lng, c.lat]
        shape = new amap.Circle({
          center,
          radius: item.radius || 300,
          strokeColor: color,
          strokeWeight: 2,
          fillColor: color,
          fillOpacity: 0.16,
          zIndex: 50,
          cursor: 'pointer',
        })
      } else {
        continue
      }
      const marker = new amap.Marker({
        position: center,
        content: mapLabelHtml(color, escapeHtml(item.name)),
        offset: new amap.Pixel(0, 0),
        zIndex: 120,
      })
      bindSelect(shape, item, center)
      bindSelect(marker, item, center)
      add(shape, true)
      add(marker)
    }
  })
}

async function reload() {
  loading.value = true
  selected.value = null
  try {
    await ensureMap()
    const { data } = await getRoomMap({
      status: filterStatus.value,
      placeType: filterPlaceType.value,
      limit: 500,
    })
    rooms.value = data?.list || []
    paintRooms()
    if (!rooms.value.length) {
      ElMessage.info('当前筛选条件下暂无聊天区')
    }
  } catch (e) {
    showApiError(e, '加载全景地图失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.dot--ok {
  background: var(--el-color-success);
}
.dot--pending {
  background: var(--el-color-warning);
}
.dot--reject {
  background: var(--el-text-color-secondary);
}
.dot--disabled {
  background: var(--el-color-danger);
}
</style>
