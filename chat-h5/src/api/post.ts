import { get, post } from './request'
import type { PostCategoryValue } from '@/utils/post/postCategory'

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

export interface PostPublisher {
  id: number
  nickname: string
  avatarUrl: string
}

export interface PostDiscussPreview {
  fromName: string
  contentType: number
  content: string
}

export interface PostItem {
  id: number
  roomId: number
  category: string
  pinned: boolean
  title: string
  content: string
  images: string[]
  imageIds?: number[]
  extras?: PostExtras
  placeName: string
  latitude: number
  longitude: number
  startAt: number
  endAt: number
  status: number
  publisherId: number
  publisher?: PostPublisher
  discussMessageCount?: number
  discussPreview?: PostDiscussPreview
  createdAt: string
  updatedAt: string
  distance: number
  signupSummary?: {
    confirmed: number
    waitlisted: number
    notCheckedIn?: number
    checkedIn?: number
  }
  mySignupStatus?: number
  myCheckedInAt?: number
}

export interface PostListParams {
  page?: number
  pageSize?: number
  latitude?: number
  longitude?: number
  roomId?: number
  category?: string
}

export interface SponsoredItem {
  campaignId: number
  title: string
  subtitle: string
  imageUrl: string
  advertiserName: string
  badge: string
  landingType: string
  landingValue: string
  distance: number
  impressionToken: string
  requestId: string
  slotIndex: number
}

export type FeedItem =
  { type: 'post'; post: PostItem } | { type: 'sponsored'; sponsored: SponsoredItem }

export interface FeedItemDto {
  type: string
  post?: PostItem
  sponsored?: SponsoredItem
}

export function normalizeFeedItem(raw: FeedItemDto | null | undefined): FeedItem | null {
  if (!raw || typeof raw !== 'object') return null
  if (raw.type === 'sponsored' && raw.sponsored) {
    return { type: 'sponsored', sponsored: raw.sponsored }
  }
  if (raw.type === 'post' && raw.post) {
    return { type: 'post', post: raw.post }
  }
  return null
}

export function feedItemKey(item: FeedItem): string {
  if (item.type === 'sponsored') {
    return `ad-${item.sponsored.requestId}-${item.sponsored.slotIndex}`
  }
  return `post-${item.post.id}`
}

export interface PostDetail extends PostItem {
  mySignupStatus: number
  myCheckedInAt?: number
}

export interface PostWritePayload {
  roomId?: number
  category: PostCategoryValue | string
  pinned?: boolean
  title: string
  content: string
  images: number[]
  extras?: PostExtras
  placeName?: string
  latitude?: number
  longitude?: number
  startAt: number
  endAt: number
}

export interface PostSignupItem {
  id: number
  userId: number
  userNickname: string
  signupStatus: number
  signupStatusText: string
  checkedInAt?: string
  createdAt: string
  cancelledAt: string
  remark?: string
}

export interface PostSignupListCounts {
  total: number
  confirmed: number
  waitlisted: number
  cancelled: number
  notCheckedIn: number
  checkedIn: number
}

export const emptySignupListCounts = (): PostSignupListCounts => ({
  total: 0,
  confirmed: 0,
  waitlisted: 0,
  cancelled: 0,
  notCheckedIn: 0,
  checkedIn: 0,
})

export interface PostSignupListParams {
  postId: number
  page?: number
  pageSize?: number
  status?: number
  checkIn?: number
}

export interface PostCheckInCode {
  postId: number
  postTitle: string
  placeName?: string
  startAt?: number
  endAt?: number
  payload: string
  checkedIn: boolean
  checkedInAt: number
  checkedInAtText?: string
  userNickname?: string
  signupAt?: string
}

export interface PostCheckInResult {
  postId: number
  postTitle: string
  userId: number
  userNickname: string
  userAvatarUrl: string
  signupAt: string
  checkedInAt: number
  checkedInAtText: string
  alreadyCheckedIn: boolean
  operatorId: number
  operatorNickname: string
  operatorAvatarUrl: string
}

export interface CheckInResultPayload {
  kind: 'success' | 'already' | 'fail'
  message?: string
  postId?: number
  postTitle?: string
  userId?: number
  userNickname?: string
  userAvatarUrl?: string
  signupAt?: string
  checkedInAtText?: string
  operatorNickname?: string
  operatorAvatarUrl?: string
}

function withListParams(params: PostListParams = {}) {
  const { page = 1, pageSize = 20, latitude, longitude, roomId, category } = params
  const query: Record<string, string | number> = { page, pageSize }
  if (latitude !== undefined && latitude !== 0) query.latitude = latitude
  if (longitude !== undefined && longitude !== 0) query.longitude = longitude
  if (roomId !== undefined && roomId > 0) query.roomId = roomId
  if (category) query.category = category
  return query
}

export function postList(params: PostListParams = {}) {
  return get<{ total: number; list: FeedItemDto[] }>('/api/post/list', withListParams(params))
}

export function postMyList(params: PostListParams = {}) {
  return get<{ total: number; list: PostItem[] }>('/api/post/my/list', withListParams(params))
}

export function postMyJoinedList(params: PostListParams = {}) {
  return get<{ total: number; list: PostItem[] }>(
    '/api/post/my/joined/list',
    withListParams(params),
  )
}

export function postUserList(userId: number, params: PostListParams = {}) {
  return get<{ total: number; list: PostItem[] }>('/api/post/user/list', {
    userId,
    ...withListParams(params),
  })
}

export function postDetail(id: number) {
  return get<PostDetail>('/api/post/detail', { id })
}

export function postCreate(data: PostWritePayload) {
  return post<{ id: number }>('/api/post/create', data)
}

export function postUpdate(data: PostWritePayload & { id: number }) {
  return post<null>('/api/post/update', data)
}

export function postWithdraw(id: number) {
  return post<null>('/api/post/withdraw', { id })
}

export function postDelete(id: number) {
  return post<null>('/api/post/delete', { id })
}

export function postJoinedRemove(id: number) {
  return post<null>('/api/post/my/joined/remove', { id })
}

export function postSignup(postId: number) {
  return post<{ status: number }>('/api/post/signup', { postId })
}

export function postSignupCancel(postId: number) {
  return post('/api/post/signup/cancel', { postId })
}

export function postCheckInCode(postId: number) {
  return get<PostCheckInCode>('/api/post/signup/check-in-code', { postId })
}

export function postCheckIn(data: {
  payload?: string
  postId?: number
  userId?: number
  token?: string
}) {
  return post<PostCheckInResult>('/api/post/signup/check-in', data)
}

export function postSignupList(params: PostSignupListParams) {
  const { postId, page = 1, pageSize = 20, status, checkIn } = params
  const query: Record<string, number> = { postId, page, pageSize }
  if (status !== undefined) query.status = status
  if (checkIn !== undefined) query.checkIn = checkIn
  return get<{
    total: number
    list: PostSignupItem[]
    needCheckIn: boolean
    counts: PostSignupListCounts
  }>('/api/post/signup/list', query)
}

export function postSignupRemark(data: { postId: number; userId: number; remark: string }) {
  return post('/api/post/signup/remark', data)
}
