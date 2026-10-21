import { get, post } from './request'
import type { MessageItem } from './message'

export interface NearbyRoomItem {
  id: number
  name: string
  placeType: number
  placeTypeText?: string
  shapeType: number
  centerLat: number
  centerLng: number
  radius: number
  distance: number
  requireGeofence: boolean
}

export interface NearbyBootstrapResult {
  rooms: NearbyRoomItem[]
  onlineCount: number
  messages?: MessageItem[]
}

export interface RoomBootstrapResult {
  room: RoomDetail
  onlineCount: number
  messages?: MessageItem[]
}

export interface RoomDetail {
  id: number
  name: string
  description: string
  placeType: number
  placeTypeText?: string
  shapeType: number
  centerLat: number
  centerLng: number
  radius: number
  requireGeofence: boolean
}

export interface RoomApplicationItem {
  id: number
  name: string
  description: string
  placeType: number
  placeTypeText?: string
  status: number
  statusText: string
  radius: number
  createdAt: string
}

export function nearbyBootstrap(
  lat: number,
  lng: number,
  limit: number,
  withMessages: boolean,
  radius: number,
) {
  return get<NearbyBootstrapResult>(
    '/api/room/nearby/bootstrap',
    {
      lat,
      lng,
      radius,
      limit,
      withMessages,
    },
    { silent: true },
  )
}

export function roomDetail(roomId: number) {
  return get<RoomDetail>(`/api/room/${roomId}`)
}

export function roomBootstrap(
  roomId: number,
  limit: number,
  withMessages: boolean,
  lat?: number,
  lng?: number,
) {
  const params: Record<string, number | boolean> = {
    roomId,
    limit,
    withMessages,
  }
  if (lat !== undefined && lat !== 0) params.lat = lat
  if (lng !== undefined && lng !== 0) params.lng = lng
  return get<RoomBootstrapResult>('/api/room/bootstrap', params, { silent: true })
}

export function createRoom(data: {
  name: string
  description?: string
  placeType: number
  shapeType: 1 | 2
  centerLat: number
  centerLng: number
  radius: number
  polygon?: [number, number][]
}) {
  return post<{ id: number }>('/api/room/create', data)
}

export function getMyRoomApplications(page = 1, pageSize = 50) {
  return get<{ total: number; list: RoomApplicationItem[] }>('/api/room/applications', {
    page,
    pageSize,
  })
}
