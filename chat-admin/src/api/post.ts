import { request } from './http'

export interface PostPublisher {
  id: number
  nickname: string
  avatarUrl?: string
}

export interface PostExtras {
  amount?: number
  negotiable?: boolean
  needSignup?: boolean
  needCheckIn?: boolean
  maxParticipants?: number
  allowWaitlist?: boolean
  signupDeadlineAt?: number
  feeType?: number
  feeDescription?: string
}

export interface PostSignupSummary {
  confirmed: number
  waitlisted: number
  cancelled?: number
  notCheckedIn?: number
  checkedIn?: number
}

export interface PostItem {
  id: number
  roomId: number
  category: string
  pinned: boolean
  title: string
  content: string
  images: string[]
  placeName: string
  latitude: number
  longitude: number
  startAt: number
  endAt: number
  status: number
  publisherId: number
  publisher?: PostPublisher
  discussMessageCount?: number
  createdAt: string
  updatedAt: string
  extras?: PostExtras
  signupSummary?: PostSignupSummary
}

export type PostDetail = PostItem

export interface PostListRequest {
  page: number
  pageSize: number
  status?: number
  publisherId?: number
  roomId?: number
  category?: string
  keyword?: string
}

export interface PostListResponse {
  total: number
  list: PostItem[]
}

export interface PostSignupListCounts {
  total: number
  confirmed: number
  waitlisted: number
  cancelled: number
  notCheckedIn: number
  checkedIn: number
}

export interface PostSignupItem {
  id: number
  userId: number
  userNickname: string
  userMobile: string
  signupStatus: number
  signupStatusText: string
  checkedInAt?: string
  createdAt: string
  cancelledAt: string
  remark?: string
}

export interface PostSignupListRequest {
  postId: number
  page: number
  pageSize: number
  status?: number
  checkIn?: number
}

export interface PostSignupListResponse {
  total: number
  list: PostSignupItem[]
  needCheckIn: boolean
  counts: PostSignupListCounts
}

export const getPostList = (params: PostListRequest) =>
  request.get<PostListResponse>('/api/admin/post/list', params)

export const getPostDetail = (id: number) =>
  request.get<PostDetail>('/api/admin/post/detail', { id })

export const updatePostStatus = (id: number, status: number) =>
  request.post('/api/admin/post/update/status', { id, status })

export const getPostSignupList = (params: PostSignupListRequest) =>
  request.get<PostSignupListResponse>('/api/admin/post/signup/list', params)

export interface PostMapRoomItem {
  roomId: number
  name: string
  placeType: number
  placeTypeText: string
  shapeType: number
  centerLat: number
  centerLng: number
  radius: number
  polygon?: [number, number][]
  roomStatus: number
  roomStatusText: string
  postCount: number
}

export interface PostMapIndependentItem {
  id: number
  title: string
  category: string
  placeName: string
  latitude: number
  longitude: number
  status: number
  createdAt: string
}

export interface PostMapRequest {
  status?: number
  category?: string
  independentLimit?: number
}

export interface PostMapResponse {
  rooms: PostMapRoomItem[]
  independent: PostMapIndependentItem[]
}

export const getPostMap = (params?: PostMapRequest) =>
  request.get<PostMapResponse>('/api/admin/post/map', params)
