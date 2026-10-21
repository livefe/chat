import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showFailToast } from 'vant'

import { searchRegion, type RegionItem } from '@/api/region'
import { useMapPoiSearch } from '@/composables/useMapPoiSearch'
import { useLocationStore } from '@/stores/location'
import { getErrorMessage } from '@/utils/apiError'
import { realPlaceNameOrEmpty } from '@/utils/map/placeName'
import { showCatchError } from '@/utils/toast'
import {
  createTiandituMap,
  formatPlaceName,
  getRegionCenter,
  hasTiandituKey,
  loadTiandituMap,
  resizeTiandituMap,
  reverseGeocodeMapCoordsDetail,
  reverseGeocodeWgs84,
  type TiandituMap,
  type TiandituPoi,
} from '@/utils/map/tianditu'

export interface MapLocationStore {
  placeName: string
  latitude: number
  longitude: number
  ready: boolean
  source: 'gps' | 'manual' | null
  setLocation: (
    next: { placeName: string; latitude: number; longitude: number },
    source?: 'gps' | 'manual',
  ) => void
  commit?: () => void
}

export function useMapLocationPicker(locationStore: MapLocationStore) {
  const router = useRouter()
  const gpsStore = useLocationStore().gps

  const mapRef = ref<HTMLDivElement | null>(null)
  const mapError = ref('')
  const mapLoading = ref(true)
  const confirming = ref(false)
  const geocoding = ref(false)
  const cityPickerVisible = ref(false)
  const selectedRegion = ref<RegionItem | null>(null)

  const centerPlaceName = ref('')
  const centerLat = ref(0)
  const centerLng = ref(0)
  const markerDragging = ref(false)
  const markerBounce = ref(false)

  let map: TiandituMap | null = null
  let geocodeTimer: ReturnType<typeof setTimeout> | null = null
  let bounceTimer: ReturnType<typeof setTimeout> | null = null
  let mapResizeObserver: ResizeObserver | null = null
  let skipNextGeocode = false

  function readMapCenter() {
    if (!map) return
    const center = map.getCenter()
    centerLat.value = center.getLat()
    centerLng.value = center.getLng()
  }

  const {
    searchKey,
    searchLoading,
    searchResults,
    showSearchList,
    onSearchInput,
    runSearch,
    clearSearch,
  } = useMapPoiSearch({
    center: () => {
      readMapCenter()
      return { lat: centerLat.value, lng: centerLng.value }
    },
    regionCode: () => selectedRegion.value?.tiandituGb || undefined,
  })

  function scheduleGeocode() {
    if (skipNextGeocode) {
      skipNextGeocode = false
      return
    }
    if (geocodeTimer) clearTimeout(geocodeTimer)
    geocodeTimer = setTimeout(() => {
      void updateCenterLabel()
    }, 350)
  }

  async function updateCenterLabel() {
    if (!hasTiandituKey()) {
      centerPlaceName.value = ''
      return
    }
    readMapCenter()
    geocoding.value = true
    try {
      const result = await reverseGeocodeWgs84(centerLat.value, centerLng.value)
      centerPlaceName.value = realPlaceNameOrEmpty(result.placeName)
    } catch {
      centerPlaceName.value = ''
    } finally {
      geocoding.value = false
    }
  }

  async function resolveInitialGeocode() {
    const geocodeDetail = await geocodeCenterDetailAndApplyLabel()
    void detectRegionFromMapCenter(geocodeDetail)
  }

  async function geocodeCenterDetailAndApplyLabel() {
    if (!hasTiandituKey()) {
      centerPlaceName.value = ''
      return null
    }
    readMapCenter()
    geocoding.value = true
    try {
      const detail = await reverseGeocodeMapCoordsDetail(centerLat.value, centerLng.value)
      centerPlaceName.value = realPlaceNameOrEmpty(formatPlaceName(detail))
      return detail
    } catch {
      centerPlaceName.value = ''
      return null
    } finally {
      geocoding.value = false
    }
  }

  function bounceMarker() {
    markerBounce.value = true
    if (bounceTimer) clearTimeout(bounceTimer)
    bounceTimer = setTimeout(() => {
      markerBounce.value = false
    }, 360)
  }

  async function panMapToRegion(region: RegionItem) {
    if (!region.tiandituGb || !map || !window.T?.LngLat) return
    const center = await getRegionCenter(region.tiandituGb, region.name)
    if (!center) {
      showFailToast('定位失败')
      return
    }

    map.panTo(new window.T.LngLat(center.lng, center.lat))
    centerLat.value = center.lat
    centerLng.value = center.lng
    await updateCenterLabel()
    bounceMarker()
  }

  async function detectRegionFromMapCenter(
    detail?: Awaited<ReturnType<typeof reverseGeocodeMapCoordsDetail>>,
  ) {
    if (!hasTiandituKey() || selectedRegion.value || !detail) return
    try {
      const comp = detail.addressComponent
      const keyword = comp?.city || comp?.county || comp?.province
      if (!keyword) return

      let level: number | undefined
      if (comp?.county) level = 3
      else if (comp?.city) level = 2
      else level = 1

      const res = await searchRegion({ keyword, level, limit: 1 })
      const region = res.data.list[0]
      if (region) {
        selectedRegion.value = region
      }
    } catch {}
  }

  async function selectRegion(region: RegionItem) {
    selectedRegion.value = region
    cityPickerVisible.value = false
    await nextTick()
    await panMapToRegion(region)
  }

  function onMapMoveStart() {
    markerDragging.value = true
    markerBounce.value = false
  }

  function onMapMoveEnd() {
    markerDragging.value = false
    readMapCenter()
    scheduleGeocode()
    bounceMarker()
  }

  async function initMap() {
    if (!mapRef.value) return
    if (!hasTiandituKey()) {
      mapError.value = '未配置 VITE_TIANDITU_TK，无法加载地图'
      mapLoading.value = false
      return
    }

    try {
      await loadTiandituMap()
    } catch (e: unknown) {
      mapError.value = getErrorMessage(e, '地图加载失败')
      mapLoading.value = false
      return
    }

    const defaultWgs = { lat: 39.90923, lng: 116.397428 }
    let wgsLat = defaultWgs.lat
    let wgsLng = defaultWgs.lng

    if (locationStore.ready) {
      wgsLat = locationStore.latitude
      wgsLng = locationStore.longitude
      centerPlaceName.value = locationStore.placeName
    } else if (gpsStore.valid) {
      wgsLat = gpsStore.lat
      wgsLng = gpsStore.lng
    }

    const startLat = wgsLat
    const startLng = wgsLng

    centerLat.value = startLat
    centerLng.value = startLng

    try {
      skipNextGeocode = true
      map = createTiandituMap(mapRef.value, startLng, startLat, 16)
      map.addEventListener('movestart', onMapMoveStart)
      map.addEventListener('moveend', onMapMoveEnd)
      if (mapRef.value && typeof ResizeObserver !== 'undefined') {
        mapResizeObserver = new ResizeObserver(() => resizeTiandituMap(map))
        mapResizeObserver.observe(mapRef.value)
      }
      const draftName = realPlaceNameOrEmpty(centerPlaceName.value)
      if (draftName) {
        centerPlaceName.value = draftName
      } else {
        void resolveInitialGeocode()
      }
    } catch (e: unknown) {
      mapError.value = getErrorMessage(e, '地图初始化失败')
    } finally {
      mapLoading.value = false
      resizeTiandituMap(map)
    }
  }

  async function useCurrentLocation() {
    await gpsStore.refresh()
    if (!gpsStore.valid) {
      showFailToast('定位失败')
      return
    }
    centerLat.value = gpsStore.lat
    centerLng.value = gpsStore.lng
    if (map && window.T?.LngLat) {
      skipNextGeocode = true
      map.panTo(new window.T.LngLat(gpsStore.lng, gpsStore.lat))
    }
    const geocodeDetail = await geocodeCenterDetailAndApplyLabel()
    selectedRegion.value = null
    await detectRegionFromMapCenter(geocodeDetail)
    bounceMarker()
  }

  function selectPoi(poi: TiandituPoi) {
    centerLat.value = poi.lat
    centerLng.value = poi.lng
    centerPlaceName.value = poi.name
    clearSearch()
    if (map && window.T?.LngLat) {
      skipNextGeocode = true
      map.panTo(new window.T.LngLat(poi.lng, poi.lat))
    }
    bounceMarker()
  }

  async function confirm() {
    if (mapError.value && !gpsStore.valid) {
      showFailToast(mapError.value)
      return
    }
    if (confirming.value) return
    confirming.value = true
    try {
      readMapCenter()
      let placeName = realPlaceNameOrEmpty(centerPlaceName.value)
      if (!placeName) {
        if (!hasTiandituKey()) {
          showFailToast('无法解析位置名称，请稍后重试')
          return
        }
        const resolved = await reverseGeocodeWgs84(centerLat.value, centerLng.value)
        placeName = realPlaceNameOrEmpty(resolved.placeName)
        if (!placeName) {
          showFailToast('无法获取该位置名称，请换个地点或重试')
          return
        }
        centerPlaceName.value = placeName
      }
      locationStore.setLocation(
        {
          placeName,
          latitude: centerLat.value,
          longitude: centerLng.value,
        },
        'manual',
      )
      locationStore.commit?.()
      router.back()
    } catch (e: unknown) {
      showCatchError(e, '定位失败', true)
    } finally {
      confirming.value = false
    }
  }

  onMounted(async () => {
    await nextTick()
    const draftReady = locationStore.ready
    const draftManual = draftReady && locationStore.source === 'manual'

    if (!draftManual && !locationStore.ready && !gpsStore.valid) {
      await gpsStore.ensureFreshLocation()
    }
    await initMap()
  })

  onBeforeUnmount(() => {
    if (geocodeTimer) clearTimeout(geocodeTimer)
    if (bounceTimer) clearTimeout(bounceTimer)
    mapResizeObserver?.disconnect()
    mapResizeObserver = null
    if (map) {
      map.removeEventListener('movestart', onMapMoveStart)
      map.removeEventListener('moveend', onMapMoveEnd)
      map = null
    }
  })

  return {
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
  }
}
