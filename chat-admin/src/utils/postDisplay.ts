import type { PostDetail, PostItem, PostSignupListCounts } from '@/api/post'
import {
  POST_CATEGORY_OPTIONS,
  POST_STATUS_OPTIONS,
  PostCategory,
  PostStatus,
  SignupStatus,
} from '@/constants/post'

export const emptySignupListCounts = (): PostSignupListCounts => ({
  total: 0,
  confirmed: 0,
  waitlisted: 0,
  cancelled: 0,
  notCheckedIn: 0,
  checkedIn: 0,
})

export function postCategoryLabel(category?: string): string {
  if (!category) return '-'
  return POST_CATEGORY_OPTIONS.find((x) => x.value === category)?.label || category
}

export function statusLabel(status: number) {
  return POST_STATUS_OPTIONS.find((o) => o.value === status)?.label ?? String(status)
}

export function statusTag(status: number): 'success' | 'warning' | 'danger' | 'info' {
  switch (status) {
    case PostStatus.Published:
      return 'success'
    case PostStatus.Pending:
      return 'warning'
    case PostStatus.Rejected:
      return 'danger'
    default:
      return 'info'
  }
}

export function formatIdlePrice(row: PostItem): string {
  const ex = row.extras
  if (!ex) return '-'
  const price = Number(ex.amount) || 0
  if (price <= 0 && !ex.negotiable) return '-'
  let text = price > 0 ? `¥${price}` : '面议'
  if (ex.negotiable && price > 0) text += ' · 可议'
  return text
}

export function formatExtraBrief(row: PostItem): string {
  const ex = row.extras
  if (!ex) return '-'
  if (row.category === PostCategory.Idle) return formatIdlePrice(row)
  if (ex.needSignup) {
    const confirmed = row.signupSummary?.confirmed ?? 0
    let text = ''
    if (ex.maxParticipants) {
      text = `报名 ${confirmed}/${ex.maxParticipants}`
    } else {
      text = confirmed > 0 ? `报名 ${confirmed}` : '需报名'
    }
    if (ex.needCheckIn) text += ' · 核验'
    return text
  }
  if (ex.feeType === 1) return `收费 ¥${ex.amount}`
  if (ex.feeType === 2) return `悬赏 ¥${ex.amount}`
  return '-'
}

export function extraTagType(row: PostItem): 'warning' | 'success' | 'primary' | 'info' {
  const ex = row.extras
  if (!ex) return 'info'
  if (row.category === PostCategory.Idle) return 'warning'
  if (ex.needSignup) return 'primary'
  if (ex.feeType === 1) return 'warning'
  if (ex.feeType === 2) return 'success'
  return 'info'
}

export function hasExtrasSection(row: PostDetail): boolean {
  const ex = row.extras
  if (!ex) return false
  if (row.category === PostCategory.Idle) {
    return !!(ex.amount || ex.negotiable)
  }
  return !!(ex.needSignup || ex.feeType === 1 || ex.feeType === 2)
}

export function signupStatusTag(status: number): 'success' | 'warning' | 'info' {
  switch (status) {
    case SignupStatus.Confirmed:
      return 'success'
    case SignupStatus.Waitlisted:
      return 'warning'
    default:
      return 'info'
  }
}

export function signupCheckInText(row: { signupStatus: number; checkedInAt?: string }): string {
  if (row.signupStatus !== SignupStatus.Confirmed) return '-'
  return row.checkedInAt ? '已核验' : '未核验'
}

export function signupCheckInTag(row: {
  signupStatus: number
  checkedInAt?: string
}): 'success' | 'warning' | 'info' {
  if (row.signupStatus !== SignupStatus.Confirmed) return 'info'
  return row.checkedInAt ? 'success' : 'info'
}
