import { PLACE_NAME_NEARBY, realPlaceNameOrEmpty } from '@/utils/map/placeName'

export type ChatLocationContent = {
  latitude: number
  longitude: number
  placeName: string
}

export function encodeChatLocation(loc: {
  latitude: number
  longitude: number
  placeName?: string
}): string {
  return JSON.stringify({
    latitude: loc.latitude,
    longitude: loc.longitude,
    placeName: realPlaceNameOrEmpty(loc.placeName) || PLACE_NAME_NEARBY,
  })
}

export function parseChatLocation(content: string | undefined): ChatLocationContent | null {
  if (!content?.trim()) return null
  try {
    const raw = JSON.parse(content) as {
      latitude?: unknown
      longitude?: unknown
      placeName?: unknown
    }
    const { latitude, longitude } = raw
    if (typeof latitude !== 'number' || typeof longitude !== 'number') return null
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) return null
    if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) return null
    if (latitude === 0 && longitude === 0) return null
    return {
      latitude,
      longitude,
      placeName:
        (typeof raw.placeName === 'string' && realPlaceNameOrEmpty(raw.placeName)) ||
        PLACE_NAME_NEARBY,
    }
  } catch {
    return null
  }
}

export function openChatLocationMap(loc: ChatLocationContent) {
  const q = encodeURIComponent(loc.placeName)
  const url = `https://uri.amap.com/marker?position=${loc.longitude},${loc.latitude}&name=${q}`
  window.open(url, '_blank', 'noopener,noreferrer')
}
