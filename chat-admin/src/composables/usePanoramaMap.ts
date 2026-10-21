import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { hasAmapKey, loadAmap, type AMapNamespace } from '@/composables/useAmapLoader'
import { clampPanoramaZoom, fitPanoramaView, panoramaMapOptions } from '@/utils/mapView'
import type { LngLatTuple } from '@/utils/mapOverlay'

type AddOverlay = (overlay: AMap.Overlay, fit?: boolean) => void

export function usePanoramaMap(options: {
  visible: () => boolean
  container: () => HTMLElement | null | undefined
  onShow: () => Promise<void>
  onHide: () => void
}) {
  const mapError = ref('')

  let amap: AMapNamespace | null = null
  let map: AMap.Map | null = null
  let overlays: AMap.Overlay[] = []
  let fitTargets: AMap.Overlay[] = []
  let infoWindow: AMap.InfoWindow | null = null

  function clearOverlays() {
    if (map && overlays.length) {
      map.remove(overlays)
    }
    overlays = []
    fitTargets = []
    infoWindow?.close()
  }

  function destroy() {
    clearOverlays()
    if (map) map.destroy()
    map = null
  }

  function openInfo(html: string, center: LngLatTuple) {
    if (!map || !amap) return
    if (!infoWindow) {
      infoWindow = new amap.InfoWindow({ offset: new amap.Pixel(0, -20) })
    }
    infoWindow.setContent(html)
    infoWindow.open(map, center)
  }

  function paint(build: (amap: AMapNamespace, add: AddOverlay) => void) {
    if (!map || !amap) return
    const activeMap = map
    clearOverlays()
    build(amap, (overlay, fit = false) => {
      overlays.push(overlay)
      if (fit) fitTargets.push(overlay)
    })
    if (overlays.length) {
      activeMap.add(overlays)
    }
    requestAnimationFrame(() => {
      if (map !== activeMap) return
      activeMap.resize()
      fitPanoramaView(activeMap, fitTargets)
    })
  }

  function fitAll() {
    if (!map || !fitTargets.length) return
    fitPanoramaView(map, fitTargets)
  }

  async function ensureMap() {
    if (map) {
      await nextTick()
      map.resize()
      return
    }
    if (!hasAmapKey()) {
      mapError.value = '未配置高德地图 Key，无法加载地图'
      ElMessage.warning('未配置 VITE_AMAP_KEY，全景地图不可用')
      return
    }
    try {
      amap = await loadAmap()
    } catch (e) {
      mapError.value = e instanceof Error ? e.message : '地图加载失败'
      ElMessage.error('地图加载失败，请检查 VITE_AMAP_KEY 配置')
      return
    }
    await nextTick()
    const el = options.container()
    if (!el) return
    const created = new amap.Map(el, panoramaMapOptions)
    created.on('zoomend', () => clampPanoramaZoom(created))
    map = created
  }

  watch(options.visible, async (v) => {
    if (!v) {
      destroy()
      options.onHide()
      mapError.value = ''
      return
    }
    await nextTick()
    await options.onShow()
  })

  onBeforeUnmount(destroy)

  return { mapError, ensureMap, paint, openInfo, fitAll }
}
