<script setup lang="ts">
import { computed, onMounted, ref, toRefs } from 'vue'
import { onBeforeRouteLeave, useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { createRoom } from '@/api/room'
import { NEARBY_RADIUS_METERS, ROOM_RADIUS_MIN_METERS } from '@/constants/nearby'
import { useLocationStore } from '@/stores/location'
import { isLocationPickNavigation, locationPickRoute } from '@/utils/map/locationPick'
import { displayPlaceName } from '@/utils/map/placeName'
import {
  ROOM_PLACE_TYPE_OPTIONS,
  roomPlaceTypeLabel,
  type RoomPlaceTypeValue,
} from '@/utils/chat/chatRoomPlaceType'

const router = useRouter()
const roomLocation = useLocationStore().room

const { name, description, placeType, radius } = toRefs(roomLocation)
const submitting = ref(false)
const typeSheetVisible = ref(false)
const typeActions = ROOM_PLACE_TYPE_OPTIONS.map((o) => ({ name: o.label, value: o.value }))

const locationLabel = computed(() => {
  if (roomLocation.ready) {
    return displayPlaceName(roomLocation.placeName, {
      isCurrentGps: roomLocation.source === 'gps',
    })
  }
  return '选择位置'
})

const placeTypeLabel = computed(() =>
  placeType.value ? roomPlaceTypeLabel(placeType.value) : '请选择类型',
)

function openLocationPicker() {
  router.push(locationPickRoute('room'))
}

function onSelectPlaceType(action: { value: RoomPlaceTypeValue }) {
  placeType.value = action.value
  typeSheetVisible.value = false
}

async function submit() {
  if (!name.value) {
    showToast('请输入聊天区名称')
    return
  }
  if (!placeType.value) {
    showToast('请选择场所类型')
    return
  }
  if (!roomLocation.ready) {
    showToast('选择位置')
    return
  }
  submitting.value = true
  try {
    await createRoom({
      name: name.value,
      description: description.value,
      placeType: placeType.value,
      shapeType: 1,
      centerLat: roomLocation.latitude,
      centerLng: roomLocation.longitude,
      radius: radius.value,
    })
    showToast('创建成功，等待审核')
    router.replace({ path: '/chat', query: { tab: 'rooms' } })
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  if (!roomLocation.ready) {
    await roomLocation.initDefault()
  }
})

onBeforeRouteLeave((to) => {
  if (isLocationPickNavigation(to) && to.query?.for === 'room') return
  if (to.name === 'RoomApplications') return
  roomLocation.clear()
})
</script>

<template>
  <SubPageLayout title="新建聊天区" full-height>
    <template #right>
      <span class="link-records" @click="router.push('/chat/room/applications')">申请记录</span>
    </template>

    <div class="create-body">
      <div class="form">
        <van-cell-group inset>
          <van-field v-model="name" label="名称" placeholder="例如：我家小区" maxlength="32" />
          <van-field
            v-model="description"
            label="描述"
            placeholder="简单介绍这个聊天区"
            maxlength="100"
            rows="2"
            autosize
            type="textarea"
          />
          <van-field
            label="类型"
            readonly
            is-link
            :model-value="placeTypeLabel"
            placeholder="请选择类型"
            :class="{ 'type-placeholder': !placeType }"
            @click="typeSheetVisible = true"
          />
          <van-field label="中心位置" readonly is-link @click="openLocationPicker">
            <template #input>
              <span class="location-value">{{ locationLabel }}</span>
            </template>
          </van-field>
          <van-field label="覆盖范围">
            <template #input>
              <van-slider
                v-model="radius"
                :min="ROOM_RADIUS_MIN_METERS"
                :max="NEARBY_RADIUS_METERS"
                :step="50"
              />
              <span class="radius-label">约 {{ radius }} 米</span>
            </template>
          </van-field>
        </van-cell-group>
        <van-button
          block
          type="primary"
          class="page-btn"
          style="margin-top: 20px"
          :loading="submitting"
          @click="submit"
        >
          提交审核
        </van-button>
        <div class="tip">提交后将进入审核，通过后即可在附近聊天区列表中看到。</div>
      </div>
    </div>

    <van-action-sheet
      v-model:show="typeSheetVisible"
      :actions="typeActions"
      cancel-text="取消"
      close-on-click-action
      @select="onSelectPlaceType"
    />
  </SubPageLayout>
</template>

<style scoped>
.link-records {
  font-size: 13px;
  color: var(--text-secondary);
}

.create-body {
  position: relative;
}
.form {
  padding: 16px 0;
}
.type-placeholder :deep(.van-field__control) {
  color: var(--van-field-placeholder-text-color, #c8c8c8);
}
.location-value {
  display: block;
  width: 100%;
  text-align: right;
  color: #323232;
}
.radius-label {
  flex-shrink: 0;
  margin-left: 12px;
  font-size: 14px;
  color: #666666;
  white-space: nowrap;
}
.tip {
  font-size: 12px;
  color: #888888;
  text-align: center;
  padding: 8px 16px 24px;
}
</style>
