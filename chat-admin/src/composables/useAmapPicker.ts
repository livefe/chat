import { onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'
import { ElMessage } from 'element-plus'

import { hasAmapKey, loadAmap, type AMapNamespace } from '@/composables/useAmapLoader'
import { hasCoordinates, locateBrowserLngLat } from '@/utils/geo'
import {
  buildNeighborOverlays,
  normalizePolygonPath,
  themeColor,
  type LngLatTuple,
  type MapRoomOverlay,
} from '@/utils/mapOverlay'

type AmapShapeType = 1 | 2

export interface AmapPickerChange {
  shapeType: AmapShapeType
  centerLat?: number
  centerLng?: number
  radius?: number
  polygon?: LngLatTuple[]
}

interface AmapPickerSource {
  shapeType: AmapShapeType
  centerLat: number
  centerLng: number
  radius: number
  polygon: LngLatTuple[]
  readonly: boolean
  overlays: MapRoomOverlay[]
}

export function useAmapPicker(
  props: Readonly<AmapPickerSource>,
  mapRef: Ref<HTMLDivElement | null>,
  onChange: (payload: AmapPickerChange) => void,
) {
  const shape = ref<AmapShapeType>(Number(props.shapeType) === 2 ? 2 : 1)
  const circleCenter = ref<LngLatTuple | null>(null)
  const circleRadius = ref<number>(props.radius || 500)
  const polygonPoints = ref<LngLatTuple[]>(props.polygon ? [...props.polygon] : [])
  const polygonHint = ref('')
  const mapError = ref('')

  let amap: AMapNamespace | null = null
  let map: AMap.Map | null = null
  let circleOverlay: AMap.Circle | null = null
  let polygonOverlay: AMap.Polygon | null = null
  let centerMarker: AMap.Marker | null = null
  let mouseTool: AMap.MouseTool | null = null
  let placeSearch: AMap.PlaceSearch | null = null
  let neighborOverlays: AMap.Overlay[] = []
  let drawHandler: ((e: AMap.MouseToolDrawEvent) => void) | null = null

  onMounted(async () => {
    if (!hasAmapKey()) {
      mapError.value = '未配置高德地图 Key，无法加载地图'
      ElMessage.warning('未配置 VITE_AMAP_KEY，聊天区地图不可用')
      return
    }
    try {
      amap = await loadAmap(['AMap.MouseTool', 'AMap.PlaceSearch'])
    } catch (e) {
      mapError.value = e instanceof Error ? e.message : '地图加载失败'
      ElMessage.error('地图加载失败，请检查 VITE_AMAP_KEY 配置')
      return
    }
    if (!mapRef.value) return
    const initialCenter = hasCoordinates(props.centerLat, props.centerLng)
      ? ([props.centerLng, props.centerLat] as LngLatTuple)
      : await locateBrowserLngLat()

    const mapOptions: AMap.MapOptions = {
      zoom: 14,
      viewMode: '2D',
    }
    if (initialCenter) {
      mapOptions.center = initialCenter
    }
    const created = new amap.Map(mapRef.value, mapOptions)
    map = created

    placeSearch = new amap.PlaceSearch({ map: created })

    requestAnimationFrame(() => {
      if (map !== created) return
      created.resize()
      paintInitialShape(created)
    })

    created.on('click', (e) => {
      if (props.readonly || shape.value !== 1) return
      circleCenter.value = [e.lnglat.getLng(), e.lnglat.getLat()]
      drawCircle()
    })
  })

  onBeforeUnmount(() => {
    stopPolygonDraw(false)
    clearNeighborOverlays()
    if (map) map.destroy()
    map = null
    mouseTool = null
  })

  function paintInitialShape(created: AMap.Map) {
    if (Number(props.shapeType) === 1 && hasCoordinates(props.centerLat, props.centerLng)) {
      circleCenter.value = [props.centerLng, props.centerLat]
      circleRadius.value = props.radius || 500
      drawCircle()
    } else if (Number(props.shapeType) === 2 && props.polygon?.length) {
      polygonPoints.value = props.polygon.map((p) => [p[0], p[1]] as LngLatTuple)
      drawPolygon()
    } else if (hasCoordinates(props.centerLat, props.centerLng)) {
      created.setCenter([props.centerLng, props.centerLat])
    }
    renderNeighborOverlays()
  }

  function clearNeighborOverlays() {
    if (!map || !neighborOverlays.length) {
      neighborOverlays = []
      return
    }
    map.remove(neighborOverlays)
    neighborOverlays = []
  }

  function renderNeighborOverlays() {
    if (!map || !amap) return
    clearNeighborOverlays()
    neighborOverlays = buildNeighborOverlays(amap, props.overlays || [])
    if (neighborOverlays.length) {
      map.add(neighborOverlays)
    }
  }

  function emitChange() {
    if (shape.value === 1) {
      onChange({
        shapeType: 1,
        centerLng: circleCenter.value?.[0] ?? 0,
        centerLat: circleCenter.value?.[1] ?? 0,
        radius: circleRadius.value,
      })
    } else {
      onChange({ shapeType: 2, polygon: [...polygonPoints.value] })
    }
  }

  function drawCircle() {
    if (!map || !amap || !circleCenter.value) return
    const stroke = themeColor('--el-color-primary', '#409eff')
    if (circleOverlay) {
      circleOverlay.setCenter(circleCenter.value)
      circleOverlay.setRadius(circleRadius.value)
    } else {
      circleOverlay = new amap.Circle({
        center: circleCenter.value,
        radius: circleRadius.value,
        strokeColor: stroke,
        strokeWeight: 2,
        fillColor: stroke,
        fillOpacity: 0.2,
      })
      map.add(circleOverlay)
    }
    if (centerMarker) {
      centerMarker.setPosition(circleCenter.value)
    } else {
      centerMarker = new amap.Marker({ position: circleCenter.value })
      map.add(centerMarker)
    }
    map.setCenter(circleCenter.value)
    emitChange()
  }

  function updateCircle() {
    if (shape.value !== 1 || !circleOverlay) {
      emitChange()
      return
    }
    circleOverlay.setRadius(circleRadius.value)
    emitChange()
  }

  function clearCircle() {
    if (circleOverlay) {
      map?.remove(circleOverlay)
      circleOverlay = null
    }
    if (centerMarker) {
      map?.remove(centerMarker)
      centerMarker = null
    }
    circleCenter.value = null
  }

  function drawPolygon() {
    if (!map || !amap) return
    if (polygonOverlay) {
      map.remove(polygonOverlay)
      polygonOverlay = null
    }
    if (!polygonPoints.value.length) return
    const stroke = props.readonly
      ? themeColor('--el-color-primary', '#409eff')
      : themeColor('--el-color-danger', '#f56c6c')
    polygonOverlay = new amap.Polygon({
      path: polygonPoints.value,
      strokeColor: stroke,
      strokeWeight: 2,
      fillColor: stroke,
      fillOpacity: 0.2,
      zIndex: 60,
    })
    map.add(polygonOverlay)
    try {
      map.setFitView([polygonOverlay], false, [40, 40, 40, 40])
    } catch {
      const first = polygonPoints.value[0]
      if (first) map.setCenter(first)
    }
    emitChange()
  }

  function stopPolygonDraw(keepOverlay = false) {
    if (mouseTool) {
      if (drawHandler) {
        mouseTool.off('draw', drawHandler)
        drawHandler = null
      }
      mouseTool.close(keepOverlay)
    }
    if (map) {
      map.setStatus({ doubleClickZoom: true })
    }
  }

  function startPolygon() {
    if (!map || !amap) {
      ElMessage.warning('地图尚未加载完成')
      return
    }
    const activeMap = map
    stopPolygonDraw(false)
    clearPolygonOverlay()
    polygonPoints.value = []
    emitChange()

    const tool = mouseTool ?? new amap.MouseTool(activeMap)
    mouseTool = tool
    activeMap.setStatus({ doubleClickZoom: false })
    polygonHint.value = '点击添加顶点，双击结束绘制'

    tool.polygon({
      strokeColor: themeColor('--el-color-danger', '#f56c6c'),
      strokeWeight: 2,
      fillColor: themeColor('--el-color-danger', '#f56c6c'),
      fillOpacity: 0.2,
    })

    drawHandler = (e: AMap.MouseToolDrawEvent) => {
      const obj = e.obj
      if (drawHandler) {
        tool.off('draw', drawHandler)
        drawHandler = null
      }
      tool.close(true)
      activeMap.setStatus({ doubleClickZoom: true })

      const path = normalizePolygonPath(obj?.getPath() ?? [])
      if (obj) activeMap.remove(obj)
      if (path.length < 3) {
        polygonHint.value = '至少需要 3 个顶点，请重新绘制'
        ElMessage.warning('多边形至少需要 3 个顶点')
        polygonOverlay = null
        polygonPoints.value = []
        emitChange()
        return
      }

      polygonPoints.value = path
      polygonHint.value = ''
      drawPolygon()
    }
    tool.on('draw', drawHandler)
  }

  function clearPolygonOverlay() {
    if (polygonOverlay) {
      map?.remove(polygonOverlay)
      polygonOverlay = null
    }
  }

  function clearPolygon() {
    stopPolygonDraw(false)
    clearPolygonOverlay()
    polygonPoints.value = []
    polygonHint.value = ''
    emitChange()
  }

  function onShapeChange() {
    stopPolygonDraw(false)
    if (shape.value === 1) {
      clearPolygon()
      polygonHint.value = ''
    } else {
      clearCircle()
      polygonHint.value = '点击「开始绘制」在地图上画多边形'
    }
    emitChange()
  }

  function doSearch(keyword: string) {
    if (!keyword || !placeSearch) return
    placeSearch.search(keyword, (_status, result) => {
      const poi = typeof result === 'string' ? undefined : result.poiList?.pois?.[0]
      if (!poi) return
      const lnglat: LngLatTuple = [poi.location.lng, poi.location.lat]
      map?.setCenter(lnglat)
      if (shape.value === 1) {
        circleCenter.value = lnglat
        drawCircle()
      }
    })
  }

  watch(
    () => props.shapeType,
    (next) => {
      const normalized: AmapShapeType = Number(next) === 2 ? 2 : 1
      if (shape.value !== normalized) shape.value = normalized
    },
  )

  watch(
    () => props.overlays,
    () => {
      renderNeighborOverlays()
    },
    { deep: true },
  )

  return {
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
  }
}
