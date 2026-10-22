import { request } from './http'

export interface RoomItem {
  id: number
  name: string
  description: string
  placeType: number
  placeTypeText: string
  shapeType: number
  centerLat: number
  centerLng: number
  radius: number
  userId: number
  employeeId: number
  source: number
  sourceText: string
  status: number
  statusText: string
  requireGeofence: boolean
  auditRemark?: string
  creatorNickname: string
  creatorMobile: string
  employeeName: string
  postCount: number
  messageCount: number
  auditedByName: string
  auditedAt: string
  createdAt: string
  updatedAt: string
}

export interface RoomDetail extends RoomItem {
  polygon?: [number, number][]
}

export interface RoomListRequest {
  page: number
  pageSize: number
  id?: number
  name?: string
  status?: number
  source?: number
  placeType?: number
  userId?: number
}

export interface RoomBatchResult {
  updated: number
  skipped: number
}

export interface RoomListResponse {
  total: number
  list: RoomItem[]
}

export interface CreateRoomRequest {
  name: string
  description?: string
  placeType: number
  shapeType: 1 | 2
  centerLat?: number
  centerLng?: number
  radius?: number
  polygon?: [number, number][]
  requireGeofence?: boolean
}

export interface UpdateRoomRequest extends CreateRoomRequest {
  id: number
}

export const getRoomList = (params: RoomListRequest) =>
  request.get<RoomListResponse>('/api/admin/room/list', params)

export const getRoomDetail = (id: number) =>
  request.get<RoomDetail>('/api/admin/room/detail', { id })

export const createRoom = (data: CreateRoomRequest) => request.post('/api/admin/room/create', data)

export const updateRoom = (data: UpdateRoomRequest) => request.post('/api/admin/room/update', data)

export const auditRoom = (id: number, approve: boolean, reason = '') =>
  request.post('/api/admin/room/audit', { id, approve, reason })

export const batchAuditRoom = (ids: number[], approve: boolean, reason = '') =>
  request.post<RoomBatchResult>('/api/admin/room/batch/audit', { ids, approve, reason })

export const updateRoomStatus = (id: number, status: number) =>
  request.post('/api/admin/room/update/status', { id, status })

export interface RoomMapItem {
  id: number
  name: string
  placeType: number
  placeTypeText: string
  shapeType: number
  centerLat: number
  centerLng: number
  radius: number
  polygon?: [number, number][]
  status: number
  statusText: string
  source: number
  sourceText: string
}

export interface RoomMapRequest {
  status?: number
  placeType?: number
  source?: number
  limit?: number
}

export interface RoomMapResponse {
  list: RoomMapItem[]
}

export const getRoomMap = (params?: RoomMapRequest) =>
  request.get<RoomMapResponse>('/api/admin/room/map', params)

export interface RoomNeighborItem {
  id: number
  name: string
  placeType: number
  placeTypeText: string
  shapeType: number
  centerLat: number
  centerLng: number
  radius: number
  status: number
  statusText: string
  distance: number
  suspect: boolean
  suspectReason: string
}

export interface RoomNeighborsRequest {
  roomId?: number
  lat?: number
  lng?: number
  radiusM?: number
  excludeId?: number
  limit?: number
}

export interface RoomNeighborsResponse {
  list: RoomNeighborItem[]
}

export const getRoomNeighbors = (params: RoomNeighborsRequest) =>
  request.get<RoomNeighborsResponse>('/api/admin/room/neighbors', params)
