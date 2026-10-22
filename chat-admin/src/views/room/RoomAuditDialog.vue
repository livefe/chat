<template>
  <el-dialog
    :model-value="visible"
    title="审核聊天区"
    width="960px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
  >
    <div v-loading="loading" class="audit-body">
      <div v-if="detail" class="audit-meta">
        <div class="audit-meta__row">
          <span class="label">名称</span>
          <span>{{ detail.name }}</span>
        </div>
        <div class="audit-meta__row">
          <span class="label">类型</span>
          <span>{{ detail.placeTypeText || '-' }}</span>
        </div>
        <div class="audit-meta__row">
          <span class="label">形状</span>
          <span>{{ Number(detail.shapeType) === 2 ? '多边形' : '圆形' }}</span>
        </div>
        <div class="audit-meta__row">
          <span class="label">来源</span>
          <span>{{ detail.sourceText || '-' }}</span>
        </div>
        <div class="audit-meta__row">
          <span class="label">申请人</span>
          <UserCell
            v-if="detail.source === RoomSource.User"
            :user-id="detail.userId"
            :nickname="detail.creatorNickname"
            :mobile="detail.creatorMobile"
          />
          <span v-else>{{ detail.employeeName || '-' }}（后台员工）</span>
        </div>
        <div class="audit-meta__row">
          <span class="label">申请时间</span>
          <span>{{ detail.createdAt }}</span>
        </div>
        <div class="audit-meta__row">
          <span class="label">中心</span>
          <span>{{ detail.centerLat.toFixed(5) }}, {{ detail.centerLng.toFixed(5) }}</span>
        </div>
        <div v-if="detail.description" class="audit-meta__row audit-meta__row--wide">
          <span class="label">描述</span>
          <span>{{ detail.description }}</span>
        </div>
        <div
          v-if="detail.status === RoomStatus.Rejected"
          class="audit-meta__row audit-meta__row--wide"
        >
          <span class="label">上次驳回</span>
          <span>
            {{ detail.auditRemark || '-' }}
            <span class="muted">
              （{{ detail.auditedByName || '-' }} · {{ detail.auditedAt || '-' }}）
            </span>
          </span>
        </div>
      </div>

      <AmapPicker
        v-if="mapReady"
        :key="`audit-map-${detail?.id || 0}-${mapPolygon.length}`"
        :shape-type="mapShapeType"
        :center-lat="mapCenter.lat"
        :center-lng="mapCenter.lng"
        :radius="detail?.radius || 500"
        :polygon="mapPolygon"
        readonly
        :overlays="mapOverlays"
      />

      <div class="neighbors">
        <div class="neighbors__title">
          附近聊天区
          <el-tag v-if="suspectCount > 0" type="warning" size="small" style="margin-left: 8px">
            疑似重复 {{ suspectCount }}
          </el-tag>
        </div>
        <el-table :data="neighbors" size="small" empty-text="附近暂无其他聊天区" max-height="240">
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
          <el-table-column prop="placeTypeText" label="类型" width="90" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'warning'" size="small">
                {{ row.statusText }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="距离" width="100">
            <template #default="{ row }">{{ formatDistance(row.distance) }}</template>
          </el-table-column>
          <el-table-column label="提示" min-width="120">
            <template #default="{ row }">
              <el-tag v-if="row.suspect" type="warning" size="small">
                {{ row.suspectReason || '疑似重复' }}
              </el-tag>
              <span v-else class="muted">—</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:visible', false)">取消</el-button>
        <el-button
          type="danger"
          :disabled="!canDecide"
          :loading="submitting"
          @click="confirmReject"
        >
          驳回
        </el-button>
        <el-button
          type="success"
          :disabled="!canDecide"
          :loading="submitting"
          @click="confirmApprove"
        >
          {{ suspectCount > 0 ? '仍要通过' : '通过' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'

import AmapPicker from '@/components/amap/AmapPicker.vue'
import UserCell from '@/components/common/UserCell.vue'
import type { MapRoomOverlay } from '@/utils/mapOverlay'
import { getRoomDetail, getRoomNeighbors, type RoomDetail, type RoomNeighborItem } from '@/api/room'
import { RoomSource, RoomStatus } from '@/constants/room'
import { showApiError } from '@/utils/apiError'
import { hasCoordinates, wgs84ToGcj02 } from '@/utils/geo'
import { confirmWithReason } from '@/utils/promptReason'

const props = defineProps<{
  visible: boolean
  roomId: number
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  approve: []
  reject: [reason: string]
}>()

const loading = ref(false)
const detail = ref<RoomDetail | null>(null)
const neighbors = ref<RoomNeighborItem[]>([])
const mapReady = ref(false)
const mapCenter = ref({ lat: 0, lng: 0 })
const mapPolygon = ref<[number, number][]>([])
const mapOverlays = ref<MapRoomOverlay[]>([])
const mapShapeType = ref<1 | 2>(1)

const suspectCount = computed(() => neighbors.value.filter((n) => n.suspect).length)
const canDecide = computed(() => !loading.value && !!detail.value)

function formatDistance(m: number) {
  if (!Number.isFinite(m)) return '-'
  if (m < 1000) return `${Math.round(m)} m`
  return `${(m / 1000).toFixed(2)} km`
}

async function confirmApprove() {
  if (props.submitting || !canDecide.value) return
  const name = detail.value?.name || '该聊天区'
  const tip =
    suspectCount.value > 0
      ? `附近有 ${suspectCount.value} 个疑似重复区域，仍要通过「${name}」？`
      : `确认通过「${name}」？`
  try {
    await ElMessageBox.confirm(tip, '审核通过', { type: 'warning' })
    emit('approve')
  } catch {}
}

async function confirmReject() {
  if (props.submitting || !canDecide.value) return
  const name = detail.value?.name || '该聊天区'
  const reason = await confirmWithReason(`确认驳回「${name}」？`, {
    title: '审核驳回',
    required: true,
    inputPlaceholder: '请输入驳回原因（申请人可见）',
  })
  if (reason !== null) emit('reject', reason)
}

async function load() {
  if (!props.roomId) return
  loading.value = true
  mapReady.value = false
  try {
    const [{ data: d }, { data: near }] = await Promise.all([
      getRoomDetail(props.roomId),
      getRoomNeighbors({ roomId: props.roomId, radiusM: 1500, limit: 30 }),
    ])
    detail.value = d
    neighbors.value = near.list || []

    const shape = Number(d.shapeType) === 2 ? 2 : 1
    mapShapeType.value = shape

    const rawPolygon = Array.isArray(d.polygon) ? d.polygon : []
    if (shape === 2 && rawPolygon.length >= 3) {
      mapPolygon.value = rawPolygon.map((pt) => {
        const lng = Number(pt[0])
        const lat = Number(pt[1])
        const p = wgs84ToGcj02(lat, lng)
        return [p.lng, p.lat] as [number, number]
      })

      const sum = mapPolygon.value.reduce(
        (acc, [lng, lat]) => ({ lng: acc.lng + lng, lat: acc.lat + lat }),
        { lng: 0, lat: 0 },
      )
      mapCenter.value = {
        lat: sum.lat / mapPolygon.value.length,
        lng: sum.lng / mapPolygon.value.length,
      }
    } else {
      mapPolygon.value = []
      if (hasCoordinates(d.centerLat, d.centerLng)) {
        const c = wgs84ToGcj02(d.centerLat, d.centerLng)
        mapCenter.value = { lat: c.lat, lng: c.lng }
      } else {
        mapCenter.value = { lat: 0, lng: 0 }
      }
      if (shape === 2) {
        mapShapeType.value = 1
      }
    }

    mapOverlays.value = (near.list || []).map((n) => {
      const c = wgs84ToGcj02(n.centerLat, n.centerLng)
      return {
        id: n.id,
        name: n.name,
        centerLat: c.lat,
        centerLng: c.lng,
        radius: n.radius || 300,
        status: n.status,
        suspect: n.suspect,
      }
    })
    mapReady.value = true
  } catch (e) {
    showApiError(e, '加载审核信息失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.visible, props.roomId] as const,
  ([visible]) => {
    if (visible && props.roomId > 0) {
      load()
    } else if (!visible) {
      detail.value = null
      neighbors.value = []
      mapOverlays.value = []
      mapReady.value = false
    }
  },
)
</script>

<style scoped>
.audit-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 200px;
}
.audit-meta {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;
  padding: 4px 0 8px;
}
.audit-meta__row {
  display: flex;
  gap: 8px;
  font-size: 13px;
  color: var(--el-text-color-primary);
}
.audit-meta__row--wide {
  grid-column: 1 / -1;
}
.audit-meta__row .label {
  flex-shrink: 0;
  width: 56px;
  color: var(--el-text-color-secondary);
}
.neighbors__title {
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}
.muted {
  color: var(--el-text-color-placeholder);
}
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
