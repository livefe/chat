export interface LocationPoint {
  placeName: string
  latitude: number
  longitude: number
}

export interface GpsReading {
  readonly valid: boolean
  readonly lat: number
  readonly lng: number
  ensureFreshLocation: () => Promise<boolean>
}

export type LocationShareTarget =
  | { type: 'nearby' }
  | { type: 'private'; peerId: number }
  | { type: 'room'; roomId: number }
  | { type: 'post'; postId: number }

export type ManualLocationSource = 'gps' | 'manual'
