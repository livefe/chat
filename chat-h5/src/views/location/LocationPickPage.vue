<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MapLocationPickerView from './components/MapLocationPickerView.vue'
import { useLocationStore } from '@/stores/location'
import { isLocationPickPurpose, type LocationPickPurpose } from '@/utils/map/locationPick'

const route = useRoute()
const router = useRouter()
const picker = useLocationStore().picker

const purpose = computed<LocationPickPurpose | null>(() => {
  const raw = route.query.for
  return isLocationPickPurpose(raw) ? raw : null
})

onMounted(() => {
  if (!purpose.value) {
    void router.replace('/chat')
  }
})

const config = {
  title: '选择位置',
  confirmText: '确认',
  dragHint: '拖动地图选择位置',
}

function bootstrap() {
  if (!purpose.value) return
  picker.init(purpose.value)
}

bootstrap()
watch(purpose, bootstrap)
</script>

<template>
  <MapLocationPickerView
    v-if="purpose"
    :title="config.title"
    :confirm-text="config.confirmText"
    :drag-hint="config.dragHint"
    :location-store="picker"
  />
</template>
