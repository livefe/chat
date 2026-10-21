import { hasCoordinates, wgs84ToGcj02 } from '@/utils/geo'

export type LngLatTuple = [number, number]

export function themeColor(name: string, fallback: string) {
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  return value || fallback
}

export function escapeHtml(s: string) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

export function wgsPolygonToGcjPath(polygon: LngLatTuple[]): LngLatTuple[] {
  return polygon.map((pt) => {
    const p = wgs84ToGcj02(Number(pt[1]), Number(pt[0]))
    return [p.lng, p.lat]
  })
}

export function pathCenter(path: LngLatTuple[]): LngLatTuple {
  let cx = 0
  let cy = 0
  for (const [lng, lat] of path) {
    cx += lng
    cy += lat
  }
  return [cx / path.length, cy / path.length]
}

function toLngLatPair(p: unknown): LngLatTuple | null {
  if (Array.isArray(p)) {
    return p.length >= 2 && typeof p[0] === 'number' && typeof p[1] === 'number'
      ? [p[0], p[1]]
      : null
  }
  const pt = p as {
    getLng?: () => number
    getLat?: () => number
    lng?: number
    lat?: number
  } | null
  if (typeof pt?.getLng === 'function' && typeof pt.getLat === 'function')
    return [pt.getLng(), pt.getLat()]
  if (typeof pt?.lng === 'number' && typeof pt.lat === 'number') return [pt.lng, pt.lat]
  return null
}

export function normalizePolygonPath(raw: unknown): LngLatTuple[] {
  if (!Array.isArray(raw) || raw.length === 0) return []

  const first: unknown = raw[0]
  const ring: unknown[] = toLngLatPair(first) === null && Array.isArray(first) ? first : raw
  const points = ring.map(toLngLatPair).filter((p): p is LngLatTuple => p !== null)

  if (points.length >= 2) {
    const a = points[0]
    const b = points[points.length - 1]
    if (a && b && a[0] === b[0] && a[1] === b[1]) points.pop()
  }
  return points.filter(
    ([lng, lat]) => Number.isFinite(lng) && Number.isFinite(lat) && !(lng === 0 && lat === 0),
  )
}

export interface MapRoomOverlay {
  id: number
  name: string

  centerLat: number
  centerLng: number
  radius: number
  status?: number
  suspect?: boolean
}

export function buildNeighborOverlays(amap: typeof AMap, items: MapRoomOverlay[]): AMap.Overlay[] {
  const overlays: AMap.Overlay[] = []
  for (const item of items) {
    if (!hasCoordinates(item.centerLat, item.centerLng)) continue
    const center: LngLatTuple = [item.centerLng, item.centerLat]
    const suspect = !!item.suspect
    const stroke = suspect
      ? themeColor('--el-color-warning', '#e6a23c')
      : themeColor('--el-text-color-secondary', '#909399')
    const circle = new amap.Circle({
      center,
      radius: item.radius || 300,
      strokeColor: stroke,
      strokeWeight: suspect ? 2 : 1,
      strokeStyle: 'dashed',
      fillColor: stroke,
      fillOpacity: suspect ? 0.18 : 0.08,
      zIndex: suspect ? 50 : 40,
    })
    const marker = new amap.Marker({
      position: center,
      content: `<div style="padding:2px 6px;border-radius:3px;background:${suspect ? 'rgba(230,162,60,0.92)' : 'rgba(96,98,102,0.85)'};color:#fff;font-size:12px;white-space:nowrap;transform:translate(-50%,-120%)">${escapeHtml(item.name)}</div>`,
      offset: new amap.Pixel(0, 0),
      zIndex: suspect ? 120 : 100,
    })
    overlays.push(circle, marker)
  }
  return overlays
}

export function mapBadgeHtml(color: string, text: string | number) {
  return `<div style="min-width:28px;height:28px;line-height:28px;text-align:center;border-radius:14px;background:${color};color:#fff;font-size:12px;font-weight:600;box-shadow:0 1px 4px rgba(0,0,0,.25);transform:translate(-50%,-50%)">${text}</div>`
}

export function mapLabelHtml(color: string, text: string, maxWidth?: number) {
  const clip = maxWidth ? `max-width:${maxWidth}px;overflow:hidden;text-overflow:ellipsis;` : ''
  return `<div style="padding:2px 6px;border-radius:3px;background:${color};color:#fff;font-size:12px;white-space:nowrap;${clip}transform:translate(-50%,-120%);box-shadow:0 1px 4px rgba(0,0,0,.2)">${text}</div>`
}

export function mapInfoHtml(title: string, detail: string) {
  return `<div style="padding:4px 2px;line-height:1.5"><b>${title}</b><br/>${detail}</div>`
}
