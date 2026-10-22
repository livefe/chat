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
  auditRemark?: string
  publisherMobile: string
  roomName: string
  deleted: boolean
  auditedByName: string
  auditedAt: string
}

export type PostDetail = PostItem

export interface PostListRequest {
  page: number
  pageSize: number
  id?: number
  publisherId?: number
  publisherKeyword?: string
  roomId?: number
  category?: string
  status?: number
  keyword?: string
  startTime?: string
  endTime?: string
  sort?: 'asc' | 'desc'
  includeDeleted?: boolean
}

export interface PostListResponse {
  total: number
  list: PostItem[]
}

export interface BatchResult {
  updated: number
  skipped: number
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

export type SignupSortField = 'createdAt' | 'checkedInAt'

export interface PostSignupListRequest {
  postId: number
  page: number
  pageSize: number
  status?: number
  checkIn?: number
  userId?: number
  keyword?: string
  sortField?: SignupSortField
  sortOrder?: 'asc' | 'desc'
}

export interface PostSignupListResponse {
  total: number
  list: PostSignupItem[]
  needCheckIn: boolean
  counts: PostSignupListCounts
}

export interface UserSignupItem {
  id: number
  postId: number
  postTitle: string
  postStatus: number
  postStatusText: string
  signupStatus: number
  signupStatusText: string
  checkedIn: boolean
  checkedInAt: string
  cancelledAt: string
  createdAt: string
}

export interface UserSignupListResponse {
  total: number
  list: UserSignupItem[]
}

export const getPostList = (params: PostListRequest) =>
  request.get<PostListResponse>('/api/admin/post/list', params)

export const getPostDetail = (id: number) =>
  request.get<PostDetail>('/api/admin/post/detail', { id })

export const updatePostStatus = (id: number, status: number, reason = '') =>
  request.post('/api/admin/post/update/status', { id, status, reason })

export const batchUpdatePostStatus = (ids: number[], status: number, reason = '') =>
  request.post<BatchResult>('/api/admin/post/batch/status', { ids, status, reason })

export const getPostSignupList = (params: PostSignupListRequest) =>
  request.get<PostSignupListResponse>('/api/admin/post/signup/list', params)

export const getUserSignups = (params: { userId: number; page: number; pageSize: number }) =>
  request.get<UserSignupListResponse>('/api/admin/post/signup/by-user', params)

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
