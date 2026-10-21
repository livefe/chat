import { ref } from 'vue'

import { hasCoordinates } from '@/utils/map/coord'
import { H5StorageKey } from '@/utils/storage'

export type GpsSource = 'gps' | 'cache' | null

const LOCATE_TIMEOUT_MS = 15000
const LOCATE_FAIL = '定位失败'

interface LocationCache {
  lat: number
  lng: number
  ts: number
}

function readCache(): LocationCache | null {
  try {
    const raw = localStorage.getItem(H5StorageKey.Location)
    if (!raw) return null
    const data = JSON.parse(raw) as LocationCache
    if (!hasCoordinates(data.lat, data.lng)) return null
    return data
  } catch {
    return null
  }
}

function writeCache(lat: number, lng: number) {
  localStorage.setItem(
    H5StorageKey.Location,
    JSON.stringify({ lat, lng, ts: Date.now() } satisfies LocationCache),
  )
}

function clearCache() {
  localStorage.removeItem(H5StorageKey.Location)
}

export function createGpsLocation() {
  const lat = ref(0)
  const lng = ref(0)

  const locating = ref(false)
  const resolved = ref(false)
  const valid = ref(false)
  const error = ref('')
  const source = ref<GpsSource>(null)
  const accuracy = ref<number | null>(null)

  let initInFlight: Promise<boolean> | null = null

  function applyGps(nextLat: number, nextLng: number, nextAccuracy: number | null) {
    lat.value = nextLat
    lng.value = nextLng
    accuracy.value = nextAccuracy
    valid.value = true
    resolved.value = true
    source.value = 'gps'
    error.value = ''
    writeCache(nextLat, nextLng)
  }

  function applyCache(cached: LocationCache): boolean {
    lat.value = cached.lat
    lng.value = cached.lng
    accuracy.value = null
    valid.value = true
    resolved.value = true
    source.value = 'cache'
    error.value = ''
    return true
  }

  function restoreCache(): boolean {
    const cached = readCache()
    return cached ? applyCache(cached) : false
  }

  function clearSession() {
    clearCache()
    lat.value = 0
    lng.value = 0
    accuracy.value = null
    valid.value = false
    resolved.value = false
    source.value = null
    error.value = ''
    locating.value = false
  }

  function getCurrentPositionOnce(): Promise<boolean> {
    return new Promise((resolve) => {
      if (!navigator.geolocation) {
        resolve(false)
        return
      }

      let settled = false
      const finish = (ok: boolean) => {
        if (settled) return
        settled = true
        window.clearTimeout(hardTimer)
        resolve(ok)
      }

      const hardTimer = window.setTimeout(() => finish(false), LOCATE_TIMEOUT_MS + 500)

      try {
        navigator.geolocation.getCurrentPosition(
          (pos) => {
            const nextLat = pos.coords.latitude
            const nextLng = pos.coords.longitude
            if (!hasCoordinates(nextLat, nextLng)) {
              finish(false)
              return
            }
            const nextAccuracy = Number.isFinite(pos.coords.accuracy) ? pos.coords.accuracy : null
            applyGps(nextLat, nextLng, nextAccuracy)
            finish(true)
          },
          () => finish(false),
          {
            enableHighAccuracy: true,
            maximumAge: 0,
            timeout: LOCATE_TIMEOUT_MS,
          },
        )
      } catch {
        finish(false)
      }
    })
  }

  async function requestGps(): Promise<boolean> {
    error.value = ''

    if (!window.isSecureContext || !navigator.geolocation) {
      error.value = LOCATE_FAIL
      if (restoreCache()) return true
      valid.value = false
      source.value = null
      resolved.value = true
      return false
    }

    const ok = await getCurrentPositionOnce()
    if (ok) return true

    if (restoreCache()) {
      error.value = ''
      return true
    }

    error.value = LOCATE_FAIL
    valid.value = false
    source.value = null
    resolved.value = true
    return false
  }

  async function init(): Promise<boolean> {
    if (initInFlight) return initInFlight
    if (valid.value && hasCoordinates(lat.value, lng.value)) return true

    locating.value = true
    resolved.value = false
    initInFlight = (async () => {
      if (restoreCache()) return true
      return requestGps()
    })().finally(() => {
      locating.value = false
      initInFlight = null
      if (!valid.value) resolved.value = true
    })

    return initInFlight
  }

  async function refresh(): Promise<boolean> {
    const snap =
      valid.value && hasCoordinates(lat.value, lng.value)
        ? { lat: lat.value, lng: lng.value, accuracy: accuracy.value }
        : null
    locating.value = true
    try {
      const ok = await requestGps()
      if (!ok && snap) {
        lat.value = snap.lat
        lng.value = snap.lng
        accuracy.value = snap.accuracy
        valid.value = true
        resolved.value = true
        source.value = 'cache'
        error.value = ''
      }
      return ok
    } finally {
      locating.value = false
    }
  }

  async function ensureFreshLocation(): Promise<boolean> {
    if (valid.value && hasCoordinates(lat.value, lng.value)) return true
    if (restoreCache()) return true
    return init()
  }

  restoreCache()

  return {
    lat,
    lng,
    locating,
    resolved,
    valid,
    error,
    source,
    init,
    refresh,
    ensureFreshLocation,
    clearSession,
  }
}
