<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import RegionCityPicker from './RegionCityPicker.vue'
import { useMapLocationPicker, type MapLocationStore } from '@/composables/useMapLocationPicker'
import MapPickerBottomCard from './MapPickerBottomCard.vue'
import MapPickerCrosshair from './MapPickerCrosshair.vue'
import MapPickerSearchPanel from './MapPickerSearchPanel.vue'

const props = defineProps<{
  title: string
  confirmText: string
  dragHint: string
  locationStore: MapLocationStore
}>()

const router = useRouter()

const {
  mapRef,
  mapError,
  mapLoading,
  confirming,
  geocoding,
  searchKey,
  searchLoading,
  searchResults,
  showSearchList,
  cityPickerVisible,
  selectedRegion,
  centerPlaceName,
  markerDragging,
  markerBounce,
  useCurrentLocation,
  onSearchInput,
  runSearch,
  selectPoi,
  selectRegion,
  confirm,
} = useMapLocationPicker(props.locationStore)

const cityLabel = computed(() => selectedRegion.value?.name || '选择城市')
const searchPlaceholder = computed(() =>
  selectedRegion.value ? `在${cityLabel.value}搜索地点` : '搜索地点（先选城市范围更准）',
)
</script>

<template>
  <div class="picker-page">
    <van-nav-bar :title="title" left-arrow :border="false" @click-left="router.back()" />

    <div class="search-bar">
      <button type="button" class="city-btn" @click="cityPickerVisible = true">
        <van-icon name="location-o" size="14" />
        <span class="city-btn__text">{{ cityLabel }}</span>
        <van-icon name="arrow-down" size="12" />
      </button>
      <van-search
        v-model="searchKey"
        :placeholder="searchPlaceholder"
        shape="round"
        class="search-bar__input"
        @update:model-value="onSearchInput"
        @search="runSearch(searchKey)"
      />
    </div>

    <RegionCityPicker v-model:show="cityPickerVisible" @select="selectRegion" />

    <div class="map-wrap">
      <div ref="mapRef" class="map" />
      <MapPickerCrosshair :dragging="markerDragging" :bounce="markerBounce" />
      <van-loading v-if="mapLoading" class="map-loading" />
      <div v-if="mapError" class="map-error">
        <p>{{ mapError }}</p>
        <p class="map-error__hint">
          请在 <code>.env.local</code> 配置
          <code>VITE_TIANDITU_TK</code>，并在天地图控制台将访问域名加入白名单后重新构建 h5 镜像
        </p>
      </div>
    </div>

    <MapPickerSearchPanel
      v-if="showSearchList"
      :loading="searchLoading"
      :results="searchResults"
      @select="selectPoi"
    />

    <MapPickerBottomCard
      :place-name="centerPlaceName"
      :drag-hint="dragHint"
      :confirm-text="confirmText"
      :geocoding="geocoding"
      :confirming="confirming"
      @use-current="useCurrentLocation"
      @confirm="confirm"
    />
  </div>
</template>

<style scoped>
.picker-page {
  height: var(--app-vh, 100dvh);
  display: flex;
  flex-direction: column;
  background: var(--page-bg);
}

.search-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px 8px;
}

.city-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 148px;
  height: 34px;
  padding: 0 12px;
  border: 1px solid var(--brand-soft-strong);
  border-radius: var(--radius-lg);
  background: var(--brand-soft);
  color: var(--brand-color-dark);
  font-size: 13px;
  font-weight: 600;
}

.city-btn__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-bar__input {
  flex: 1;
  min-width: 0;
  padding: 0;
  border-radius: 999px;
  overflow: hidden;
  background: #fff;
}

.search-bar__input :deep(.van-search) {
  padding: 0;
  background: #fff;
}

.search-bar__input :deep(.van-search__content) {
  background: #fff;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 999px;
}

.search-bar__input :deep(.van-field__control) {
  background: transparent;
}

.map-wrap {
  position: relative;
  flex: 1;
  min-height: 240px;
  overflow: hidden;
}

.map {
  width: 100%;
  height: 100%;
}

.map-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.72);
  z-index: 1100;
}

.map-error {
  position: absolute;
  inset: 0;
  z-index: 1100;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px;
  text-align: center;
  background: #f7f7f7;
  color: #666666;
  font-size: 14px;
}

.map-error__hint {
  margin: 0;
  font-size: 12px;
  color: #999999;
}

.map-error code {
  color: var(--brand-color);
}
</style>
