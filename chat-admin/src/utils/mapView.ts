const PANORAMA_MIN_ZOOM = 10
const PANORAMA_MAX_ZOOM = 19

export const panoramaMapOptions = {
  zoom: 12,
  viewMode: '2D' as const,
  zooms: [PANORAMA_MIN_ZOOM, PANORAMA_MAX_ZOOM] as [number, number],
  dragEnable: true,
  zoomEnable: true,
}

export function clampPanoramaZoom(map: AMap.Map | null) {
  if (!map) return
  const z = Number(map.getZoom())
  if (!Number.isFinite(z)) return
  if (z < PANORAMA_MIN_ZOOM) map.setZoom(PANORAMA_MIN_ZOOM)
  if (z > PANORAMA_MAX_ZOOM) map.setZoom(PANORAMA_MAX_ZOOM)
}

export function fitPanoramaView(map: AMap.Map | null, overlays: AMap.Overlay[]) {
  if (!map || !overlays.length) return
  map.setFitView(overlays, false, [48, 48, 48, 48])
  clampPanoramaZoom(map)
}
