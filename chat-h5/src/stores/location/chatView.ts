import { computed, ref } from 'vue'

import { hasCoordinates } from '@/utils/map/coord'
import { displayPlaceName, realPlaceNameOrEmpty } from '@/utils/map/placeName'
import { hasTiandituKey, reverseGeocodeWgs84 } from '@/utils/map/tianditu'
import { H5StorageKey } from '@/utils/storage'

import type { GpsReading, LocationPoint } from './types'

export type ChatViewMode = 'gps' | 'browse'

interface PersistData {
  mode: ChatViewMode
  placeName: string
  gpsPlaceName: string
  browseLat: number
  browseLng: number
  committedLat: number
  committedLng: number
}

function loadPersist(): PersistData | null {
  try {
    const raw = localStorage.getItem(H5StorageKey.ChatViewLocation)
    if (!raw) return null
    const data = JSON.parse(raw) as PersistData
    if (data.mode !== 'gps' && data.mode !== 'browse') return null
    if (typeof data.placeName !== 'string' || typeof data.gpsPlaceName !== 'string') return null
    if (!hasCoordinates(data.committedLat, data.committedLng)) return null
    if (data.mode === 'browse' && !hasCoordinates(data.browseLat, data.browseLng)) return null
    return data
  } catch {
    return null
  }
}

function savePersist(data: PersistData) {
  localStorage.setItem(H5StorageKey.ChatViewLocation, JSON.stringify(data))
}

export function createChatViewLocation(gps: GpsReading) {
  const persisted = loadPersist()
  const mode = ref<ChatViewMode>(persisted?.mode ?? 'gps')
  const placeName = ref(realPlaceNameOrEmpty(persisted?.placeName))
  const browseLat = ref(persisted?.browseLat ?? 0)
  const browseLng = ref(persisted?.browseLng ?? 0)
  const committedLat = ref(persisted?.committedLat ?? 0)
  const committedLng = ref(persisted?.committedLng ?? 0)
  const gpsPlaceName = ref(realPlaceNameOrEmpty(persisted?.gpsPlaceName))
  const gpsResolving = ref(false)

  let gpsResolveSeq = 0

  const isBrowsing = computed(
    () => mode.value === 'browse' && hasCoordinates(browseLat.value, browseLng.value),
  )

  const canCommitLocation = computed(() => {
    if (isBrowsing.value) return true
    return gps.valid
  })

  const hasCommittedLocation = computed(() =>
    hasCoordinates(committedLat.value, committedLng.value),
  )

  const activeLat = computed(() => committedLat.value)
  const activeLng = computed(() => committedLng.value)

  const activePlaceName = computed(() => {
    if (!hasCommittedLocation.value) return ''
    if (isBrowsing.value) return realPlaceNameOrEmpty(placeName.value)
    return realPlaceNameOrEmpty(gpsPlaceName.value)
  })

  const headerPlaceLabel = computed(() => {
    if (!hasCommittedLocation.value) return ''
    return displayPlaceName(activePlaceName.value, { isCurrentGps: !isBrowsing.value })
  })

  const viewKey = computed(() => `${mode.value}:${activeLat.value}:${activeLng.value}`)

  function persist() {
    savePersist({
      mode: mode.value,
      placeName: realPlaceNameOrEmpty(placeName.value),
      gpsPlaceName: realPlaceNameOrEmpty(gpsPlaceName.value),
      browseLat: browseLat.value,
      browseLng: browseLng.value,
      committedLat: committedLat.value,
      committedLng: committedLng.value,
    })
  }

  async function resolveGpsPlaceName() {
    if (!hasCoordinates(committedLat.value, committedLng.value) || !hasTiandituKey()) {
      gpsPlaceName.value = ''
      gpsResolving.value = false
      persist()
      return
    }
    if (realPlaceNameOrEmpty(gpsPlaceName.value)) {
      gpsResolving.value = false
      return
    }
    const seq = ++gpsResolveSeq
    const lat = committedLat.value
    const lng = committedLng.value
    gpsResolving.value = true
    try {
      const result = await reverseGeocodeWgs84(lat, lng)
      if (seq !== gpsResolveSeq) return
      if (Math.abs(committedLat.value - lat) > 1e-6 || Math.abs(committedLng.value - lng) > 1e-6) {
        return
      }
      gpsPlaceName.value = realPlaceNameOrEmpty(result.placeName)
      persist()
    } catch {
      if (seq !== gpsResolveSeq) return
      gpsPlaceName.value = ''
      persist()
    } finally {
      if (seq === gpsResolveSeq) gpsResolving.value = false
    }
  }

  function commitActiveLocation(): boolean {
    let nextLat: number
    let nextLng: number
    if (isBrowsing.value) {
      if (!hasCoordinates(browseLat.value, browseLng.value)) return false
      nextLat = browseLat.value
      nextLng = browseLng.value
    } else {
      if (!gps.valid) return false
      nextLat = gps.lat
      nextLng = gps.lng
    }
    const coordsChanged =
      !hasCoordinates(committedLat.value, committedLng.value) ||
      Math.abs(committedLat.value - nextLat) > 1e-6 ||
      Math.abs(committedLng.value - nextLng) > 1e-6
    committedLat.value = nextLat
    committedLng.value = nextLng
    if (!isBrowsing.value && coordsChanged) {
      gpsPlaceName.value = ''
      gpsResolving.value = true
    }
    persist()
    if (!isBrowsing.value && coordsChanged) {
      void resolveGpsPlaceName()
    }
    return true
  }

  async function relocateToGps(): Promise<boolean> {
    mode.value = 'gps'
    if (!gps.valid) {
      gpsResolving.value = false
      persist()
      return false
    }
    committedLat.value = gps.lat
    committedLng.value = gps.lng
    gpsPlaceName.value = ''
    gpsResolving.value = true
    persist()
    await resolveGpsPlaceName()
    return true
  }

  function setBrowseLocation(next: LocationPoint) {
    if (!hasCoordinates(next.latitude, next.longitude)) return
    gpsResolveSeq += 1
    gpsResolving.value = false
    placeName.value = realPlaceNameOrEmpty(next.placeName)
    browseLat.value = next.latitude
    browseLng.value = next.longitude
    mode.value = 'browse'
    committedLat.value = next.latitude
    committedLng.value = next.longitude
    persist()
  }

  function clearSession() {
    gpsResolveSeq += 1
    gpsResolving.value = false
    mode.value = 'gps'
    placeName.value = ''
    gpsPlaceName.value = ''
    browseLat.value = 0
    browseLng.value = 0
    committedLat.value = 0
    committedLng.value = 0
    localStorage.removeItem(H5StorageKey.ChatViewLocation)
  }

  function ensureGpsPlaceResolved() {
    if (isBrowsing.value || !hasCommittedLocation.value) return
    if (realPlaceNameOrEmpty(gpsPlaceName.value)) return
    void resolveGpsPlaceName()
  }

  function committedCoords(): { lat: number; lng: number } | null {
    if (!hasCoordinates(committedLat.value, committedLng.value)) return null
    return { lat: committedLat.value, lng: committedLng.value }
  }

  function committedPoint(): (LocationPoint & { browsing: boolean }) | null {
    if (!hasCoordinates(committedLat.value, committedLng.value)) return null
    return {
      placeName: activePlaceName.value,
      latitude: committedLat.value,
      longitude: committedLng.value,
      browsing: isBrowsing.value,
    }
  }

  function browsePoint(): LocationPoint | null {
    if (!isBrowsing.value) return null
    return { placeName: placeName.value, latitude: browseLat.value, longitude: browseLng.value }
  }

  return {
    isBrowsing,
    canCommitLocation,
    hasCommittedLocation,
    headerPlaceLabel,
    gpsResolving,
    viewKey,
    commitActiveLocation,
    relocateToGps,
    setBrowseLocation,
    clearSession,
    ensureGpsPlaceResolved,
    committedCoords,
    committedPoint,
    browsePoint,
  }
}
