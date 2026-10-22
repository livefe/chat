<template>
  <PanoramaMapShell
    ref="shellRef"
    v-model:visible="visible"
    title="动态全景地图"
    :loading="loading"
    :error="mapError"
    :side-width="240"
  >
    <template #toolbar>
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 120px">
        <el-option
          v-for="item in POST_STATUS_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
      <el-select v-model="filterCategory" placeholder="分类" clearable style="width: 120px">
        <el-option
          v-for="item in POST_CATEGORY_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
      <el-button type="primary" @click="reload">刷新</el-button>
      <el-button @click="fitAll">适应全部</el-button>
      <span class="map-toolbar__meta">
        聊天区 {{ rooms.length }} · 独立点位 {{ indepClusters.length }}（{{ independent.length }}
        条）· 缩小到最大视野后只能拖动
      </span>
      <div class="map-legend">
        <span><i class="dot dot--room" />聊天区（数字=动态数）</span>
        <span><i class="dot dot--indep" />独立动态（同点聚合）</span>
      </div>
    </template>

    <template #side>
      <template v-if="selectedRoom">
        <div class="map-side__title">{{ selectedRoom.name }}</div>
        <div class="map-side__row"><span>聊天区 ID</span>{{ selectedRoom.roomId }}</div>
        <div class="map-side__row"><span>类型</span>{{ selectedRoom.placeTypeText || '-' }}</div>
        <div class="map-side__row"><span>区状态</span>{{ selectedRoom.roomStatusText }}</div>
        <div class="map-side__row"><span>动态数</span>{{ selectedRoom.postCount }}</div>
        <el-button type="primary" class="map-side__btn" @click="onFilterRoom">
          查看该区动态列表
        </el-button>
      </template>
      <template v-else-if="selectedCluster">
        <div class="map-side__title">
          {{ selectedCluster.placeName || '独立动态' }}
          <el-tag size="small" type="warning" style="margin-left: 6px">
            {{ selectedCluster.posts.length }} 条
          </el-tag>
        </div>
        <div class="map-side__list">
          <button
            v-for="p in selectedCluster.posts"
            :key="p.id"
            type="button"
            class="map-side__post"
            @click="emitOpenPost(p.id)"
          >
            <div class="map-side__post-title">{{ p.title }}</div>
            <div class="map-side__post-meta">
              #{{ p.id }} · {{ postCategoryLabel(p.category) }} · {{ statusLabel(p.status) }}
            </div>
          </button>
        </div>
      </template>
      <div v-else class="map-side__empty">
        蓝色为挂聊天区动态的聚合；橙色为独立动态，同一坐标会合并成一个点并显示数量。
      </div>
    </template>
  </PanoramaMapShell>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getPostMap, type PostMapIndependentItem, type PostMapRoomItem } from '@/api/post'
import PanoramaMapShell from '@/components/PanoramaMapShell.vue'
import { usePanoramaMap } from '@/composables/usePanoramaMap'
import { POST_CATEGORY_OPTIONS, POST_STATUS_OPTIONS } from '@/constants/post'
import { showApiError } from '@/utils/apiError'
import { hasCoordinates, wgs84ToGcj02 } from '@/utils/geo'
import {
  escapeHtml,
  mapBadgeHtml,
  mapInfoHtml,
  mapLabelHtml,
  pathCenter,
  themeColor,
  wgsPolygonToGcjPath,
  type LngLatTuple,
} from '@/utils/mapOverlay'
import { postCategoryLabel, statusLabel } from '@/utils/postDisplay'

interface IndepCluster {
  key: string
  lat: number
  lng: number
  placeName: string
  posts: PostMapIndependentItem[]
}

const visible = defineModel<boolean>('visible', { required: true })
const emit = defineEmits<{
  'filter-room': [roomId: number]
  'open-post': [postId: number]
}>()

const shellRef = ref<InstanceType<typeof PanoramaMapShell>>()
const loading = ref(false)
const rooms = ref<PostMapRoomItem[]>([])
const independent = ref<PostMapIndependentItem[]>([])
const selectedRoom = ref<PostMapRoomItem | null>(null)
const selectedCluster = ref<IndepCluster | null>(null)
const filterStatus = ref<number | undefined>(undefined)
const filterCategory = ref<string | undefined>(undefined)

const { mapError, ensureMap, paint, openInfo, fitAll } = usePanoramaMap({
  visible: () => visible.value,
  container: () => shellRef.value?.canvas,
  onShow: reload,
  onHide: () => {
    rooms.value = []
    independent.value = []
    selectedRoom.value = null
    selectedCluster.value = null
  },
})

function coordKey(lat: number, lng: number) {
  return `${lat.toFixed(4)},${lng.toFixed(4)}`
}

const indepClusters = computed(() => {
  const map = new Map<string, IndepCluster>()
  for (const item of independent.value) {
    if (!hasCoordinates(item.latitude, item.longitude)) continue
    const key = coordKey(item.latitude, item.longitude)
    let cluster = map.get(key)
    if (!cluster) {
      cluster = {
        key,
        lat: item.latitude,
        lng: item.longitude,
        placeName: item.placeName || '',
        posts: [],
      }
      map.set(key, cluster)
    }
    cluster.posts.push(item)
    if (!cluster.placeName && item.placeName) {
      cluster.placeName = item.placeName
    }
  }
  return Array.from(map.values())
})

function paintMap() {
  paint((amap, add) => {
    const roomColor = themeColor('--el-color-primary', '#409eff')
    const indepColor = themeColor('--el-color-warning', '#e6a23c')

    for (const item of rooms.value) {
      let center: LngLatTuple | null = null
      if (item.shapeType === 2 && item.polygon && item.polygon.length >= 3) {
        const path = wgsPolygonToGcjPath(item.polygon)
        const shapeCenter = pathCenter(path)
        center = shapeCenter
        const polygon = new amap.Polygon({
          path,
          strokeColor: roomColor,
          strokeWeight: 2,
          fillColor: roomColor,
          fillOpacity: 0.14,
          zIndex: 40,
          cursor: 'pointer',
        })
        polygon.on('click', () => selectRoom(item, shapeCenter))
        add(polygon, true)
      } else if (hasCoordinates(item.centerLat, item.centerLng)) {
        const c = wgs84ToGcj02(item.centerLat, item.centerLng)
        const shapeCenter: LngLatTuple = [c.lng, c.lat]
        center = shapeCenter
        const circle = new amap.Circle({
          center: shapeCenter,
          radius: item.radius || 300,
          strokeColor: roomColor,
          strokeWeight: 2,
          fillColor: roomColor,
          fillOpacity: 0.12,
          zIndex: 40,
          cursor: 'pointer',
        })
        circle.on('click', () => selectRoom(item, shapeCenter))
        add(circle, true)
      }
      if (!center) continue
      const markerCenter = center
      const marker = new amap.Marker({
        position: markerCenter,
        content: mapBadgeHtml(roomColor, item.postCount),
        offset: new amap.Pixel(0, 0),
        zIndex: 110,
      })
      marker.on('click', () => selectRoom(item, markerCenter))
      add(marker)
    }

    for (const cluster of indepClusters.value) {
      const c = wgs84ToGcj02(cluster.lat, cluster.lng)
      const center: LngLatTuple = [c.lng, c.lat]
      const count = cluster.posts.length
      const content =
        count > 1
          ? mapBadgeHtml(indepColor, count)
          : mapLabelHtml(
              indepColor,
              escapeHtml(
                cluster.posts[0]?.title || cluster.placeName || `#${cluster.posts[0]?.id}`,
              ),
              140,
            )
      const marker = new amap.Marker({
        position: center,
        content,
        offset: new amap.Pixel(0, 0),
        zIndex: 120,
      })
      marker.on('click', () => selectCluster(cluster, center))
      add(marker, true)
    }
  })
}

function selectRoom(item: PostMapRoomItem, center: LngLatTuple) {
  selectedRoom.value = item
  selectedCluster.value = null
  openInfo(
    mapInfoHtml(
      escapeHtml(item.name),
      `动态 ${item.postCount} 条 · ${escapeHtml(item.placeTypeText || '')}`,
    ),
    center,
  )
}

function selectCluster(cluster: IndepCluster, center: LngLatTuple) {
  selectedCluster.value = cluster
  selectedRoom.value = null
  openInfo(
    mapInfoHtml(escapeHtml(cluster.placeName || '独立动态'), `同点 ${cluster.posts.length} 条动态`),
    center,
  )
}

function onFilterRoom() {
  if (!selectedRoom.value) return
  emit('filter-room', selectedRoom.value.roomId)
  visible.value = false
}

function emitOpenPost(postId: number) {
  emit('open-post', postId)
  visible.value = false
}

async function reload() {
  loading.value = true
  selectedRoom.value = null
  selectedCluster.value = null
  try {
    await ensureMap()
    const { data } = await getPostMap({
      status: filterStatus.value,
      category: filterCategory.value,
      independentLimit: 500,
    })
    rooms.value = data?.rooms || []
    independent.value = data?.independent || []
    paintMap()
    if (!rooms.value.length && !independent.value.length) {
      ElMessage.info('当前筛选条件下暂无动态分布')
    }
  } catch (e) {
    showApiError(e, '加载动态全景地图失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.dot--room {
  background: var(--el-color-primary);
}
.dot--indep {
  background: var(--el-color-warning);
}

.map-side__btn {
  width: 100%;
  margin-top: 8px;
}

.map-side__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 56vh;
  overflow: auto;
}

.map-side__post {
  display: block;
  width: 100%;
  text-align: left;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  background: var(--el-fill-color-blank);
  padding: 8px 10px;
  cursor: pointer;
}

.map-side__post:hover {
  border-color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
}

.map-side__post-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  word-break: break-all;
}

.map-side__post-meta {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
