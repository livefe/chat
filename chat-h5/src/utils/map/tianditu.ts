const API_BASE = 'https://api.tianditu.gov.cn'

export interface TiandituPoi {
  name: string
  address: string
  lat: number
  lng: number
}

interface GeocodeAddressComponent {
  province?: string
  city?: string
  county?: string
  town?: string
  road?: string
  address?: string
  poi?: string
}

interface GeocodeResult {
  formatted_address?: string
  addressComponent?: GeocodeAddressComponent
}

let scriptPromise: Promise<void> | null = null

function getTk(): string {
  return import.meta.env.VITE_TIANDITU_TK || ''
}

export function hasTiandituKey(): boolean {
  return !!getTk()
}

export function loadTiandituMap(): Promise<void> {
  const tk = getTk()
  if (!tk) return Promise.reject(new Error('未配置 VITE_TIANDITU_TK'))
  if (typeof window !== 'undefined' && window.T?.Map) return Promise.resolve()
  if (scriptPromise) return scriptPromise

  scriptPromise = new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.type = 'text/javascript'
    script.src = `${API_BASE}/api?v=4.0&tk=${encodeURIComponent(tk)}`
    script.onload = () => {
      if (window.T?.Map) resolve()
      else reject(new Error('天地图 API 未就绪'))
    }
    script.onerror = () => reject(new Error('天地图脚本加载失败'))
    document.head.appendChild(script)
  })

  return scriptPromise
}

async function tiandituGet<T>(path: string, params: Record<string, string>): Promise<T> {
  const tk = getTk()
  if (!tk) throw new Error('未配置 VITE_TIANDITU_TK')

  const query = new URLSearchParams({ ...params, tk })
  const res = await fetch(`${API_BASE}${path}?${query.toString()}`)
  if (!res.ok) throw new Error('天地图服务请求失败')
  const data = (await res.json()) as T & {
    status?: string | number | { infocode?: string | number; cndesc?: string }
    msg?: string
    remark?: string
  }

  if (!isTiandituSuccess(data.status)) {
    const status = data.status
    const message =
      typeof status === 'object'
        ? status.cndesc || data.msg || data.remark
        : data.msg || data.remark
    throw new Error(message || '天地图服务返回错误')
  }
  return data
}

function isTiandituSuccess(status: unknown): boolean {
  if (status === undefined || status === null) return true
  if (typeof status === 'object') {
    const code = Number((status as { infocode?: string | number }).infocode ?? 0)
    return code === 0 || code === 1000
  }
  const text = String(status)
  return text === '' || text === '0' || text === '1000' || text === '200'
}

export function formatPlaceName(result?: GeocodeResult | null): string {
  if (!result) return ''
  const poi = result.addressComponent?.poi?.trim()
  if (poi) return poi
  return result.formatted_address?.trim() || ''
}

export async function reverseGeocodeMapCoordsDetail(
  mapLat: number,
  mapLng: number,
): Promise<GeocodeResult | null> {
  const postStr = JSON.stringify({ lon: mapLng, lat: mapLat, ver: 1 })
  const data = await tiandituGet<{ result?: GeocodeResult }>('/geocoder', {
    postStr,
    type: 'geocode',
  })
  return data.result ?? null
}

export async function reverseGeocodeWgs84(
  lat: number,
  lng: number,
): Promise<{ placeName: string; lat: number; lng: number }> {
  const detail = await reverseGeocodeMapCoordsDetail(lat, lng)
  return { placeName: formatPlaceName(detail), lat, lng }
}

function parseLngLatPair(lonlat?: string): { lat: number; lng: number } | null {
  if (!lonlat) return null
  const parts = lonlat.split(',')
  if (parts.length < 2) return null
  const lng = Number(parts[0])
  const lat = Number(parts[1])
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null
  return { lat, lng }
}

function parseTiandituPois(
  pois?: Array<{ name?: string; address?: string; lonlat?: string }>,
): TiandituPoi[] {
  if (!pois?.length) return []
  return pois
    .map((poi) => {
      const ll = parseLngLatPair(poi.lonlat)
      if (!ll) return null
      return {
        name: poi.name?.trim() || '未知地点',
        address: poi.address?.trim() || '',
        lat: ll.lat,
        lng: ll.lng,
      }
    })
    .filter((x): x is TiandituPoi => !!x)
}

interface AdminDistrict {
  center?: string
  bound?: string
}

function centerFromBound(bound?: string): { lat: number; lng: number } | null {
  if (!bound) return null
  const [minLng, minLat, maxLng, maxLat] = bound.split(',').map((x) => Number(x.trim()))
  if (
    minLng === undefined ||
    minLat === undefined ||
    maxLng === undefined ||
    maxLat === undefined ||
    [minLng, minLat, maxLng, maxLat].some((n) => !Number.isFinite(n))
  ) {
    return null
  }
  return {
    lng: (minLng + maxLng) / 2,
    lat: (minLat + maxLat) / 2,
  }
}

async function fetchAdminDistrict(tiandituGb: string): Promise<AdminDistrict | null> {
  const data = await tiandituGet<{
    data?: { district?: AdminDistrict[] }
    district?: AdminDistrict[]
  }>('/v2/administrative', {
    keyword: tiandituGb,
    childLevel: '0',
    extensions: 'true',
  })

  const district = data.data?.district ?? data.district
  return district?.[0] ?? null
}

export async function searchPoiInRegion(
  keyword: string,
  tiandituGb: string,
  count = 20,
): Promise<TiandituPoi[]> {
  const keyWord = keyword.trim()
  if (!keyWord || !tiandituGb) return []

  const postStr = JSON.stringify({
    keyWord,
    queryType: 12,
    specify: tiandituGb,
    start: 0,
    count,
    show: 2,
  })

  const data = await tiandituGet<{
    pois?: Array<{ name?: string; address?: string; lonlat?: string }>
  }>('/v2/search', {
    postStr,
    type: 'query',
  })

  return parseTiandituPois(data.pois)
}

export async function getRegionCenter(
  tiandituGb: string,
  regionName = '',
): Promise<{ lat: number; lng: number } | null> {
  if (!tiandituGb) return null

  try {
    const district = await fetchAdminDistrict(tiandituGb)
    const center = parseLngLatPair(district?.center) ?? centerFromBound(district?.bound)
    if (center) return center
  } catch {}

  if (!regionName) return null

  try {
    const postStr = JSON.stringify({
      keyWord: regionName,
      queryType: 12,
      specify: tiandituGb,
      start: 0,
      count: 1,
      show: 2,
    })
    const data = await tiandituGet<{
      pois?: Array<{ lonlat?: string }>
    }>('/v2/search', {
      postStr,
      type: 'query',
    })
    const poi = data.pois?.[0]
    return parseLngLatPair(poi?.lonlat)
  } catch {
    return null
  }
}

export async function searchPoiNearby(
  keyword: string,
  centerLat: number,
  centerLng: number,
  count = 10,
): Promise<TiandituPoi[]> {
  const keyWord = keyword.trim()
  if (!keyWord) return []

  const postStr = JSON.stringify({
    keyWord,
    level: 16,
    queryRadius: 5000,
    pointLonlat: `${centerLng},${centerLat}`,
    queryType: 3,
    start: 0,
    count,
  })

  const data = await tiandituGet<{
    pois?: Array<{ name?: string; address?: string; lonlat?: string }>
  }>('/v2/search', {
    postStr,
    type: 'query',
  })

  return parseTiandituPois(data.pois)
}

declare global {
  interface Window {
    T?: {
      Map: new (el: HTMLElement | string, opts?: Record<string, unknown>) => TiandituMap
      LngLat: new (lng: number, lat: number) => TiandituLngLat
    }
  }
}

interface TiandituLngLat {
  getLat(): number
  getLng(): number
}

export interface TiandituMap {
  centerAndZoom(lnglat: TiandituLngLat, zoom: number): void
  getCenter(): TiandituLngLat
  getZoom(): number
  addEventListener(type: string, handler: () => void): void
  removeEventListener(type: string, handler: () => void): void
  panTo(lnglat: TiandituLngLat): void
  checkResize?(): void
}

export function createTiandituMap(
  el: HTMLElement,
  lng: number,
  lat: number,
  zoom = 16,
): TiandituMap {
  if (!window.T?.Map || !window.T.LngLat) {
    throw new Error('天地图 API 未加载')
  }

  const map = new window.T.Map(el)
  map.centerAndZoom(new window.T.LngLat(lng, lat), zoom)
  map.checkResize?.()
  return map
}

export function resizeTiandituMap(map: TiandituMap | null) {
  map?.checkResize?.()
}
