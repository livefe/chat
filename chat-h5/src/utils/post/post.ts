import { formatConversationTime } from '@/utils/chat/chatTime'
import { PostFeeType, PostStatus, SignupStatus } from '@/constants/post'
import { pad } from '@/utils/datetime'

function formatUnixDate(d: Date): string {
  const now = new Date()
  if (d.getFullYear() === now.getFullYear()) {
    return `${d.getMonth() + 1}月${d.getDate()}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  return `${d.getFullYear()}年${pad(d.getMonth() + 1)}月${pad(d.getDate())}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatPostCreatedAt(v: string): string {
  if (!v) return ''
  const ms = new Date(v.replace(' ', 'T')).getTime()
  if (!ms || Number.isNaN(ms)) return v
  return formatConversationTime(Math.floor(ms / 1000))
}

export function formatPostScheduleRange(
  startAt: number,
  endAt: number,
): { start: string; end: string } | null {
  if (!startAt && !endAt) return null
  return {
    start: startAt ? formatUnixDate(new Date(startAt * 1000)) : '',
    end: endAt ? formatUnixDate(new Date(endAt * 1000)) : '',
  }
}

export function hasPostFee(rule: { feeType?: number }): boolean {
  return rule.feeType === PostFeeType.Charge || rule.feeType === PostFeeType.Reward
}

export function formatPostFeeBrief(rule: { feeType?: number; amount?: number }): string {
  const amount = rule.amount ?? 0
  if (rule.feeType === PostFeeType.Charge) return `收费 ¥${amount}`
  if (rule.feeType === PostFeeType.Reward) return `报酬 ¥${amount}`
  return ''
}

function formatSignupBrief(
  rule: {
    needSignup?: boolean
    maxParticipants?: number
    allowWaitlist?: boolean
  },
  summary?: { confirmed: number; waitlisted: number },
): string {
  if (!rule.needSignup) return ''
  const confirmed = summary?.confirmed ?? 0
  const waitlisted = summary?.waitlisted ?? 0
  const max = rule.maxParticipants ?? 0
  if (max > 0) {
    let text = `${confirmed}/${max}人报名`
    if (waitlisted > 0) text += ` · ${waitlisted}人候补`
    return text
  }
  if (confirmed > 0 && waitlisted > 0) return `${confirmed}人报名 · ${waitlisted}人候补`
  if (confirmed > 0) return `${confirmed}人报名`
  if (waitlisted > 0) return `${waitlisted}人候补`
  return '需报名'
}

export function formatSignupPeopleDetail(
  rule: {
    needSignup?: boolean
    maxParticipants?: number
    allowWaitlist?: boolean
  },
  summary?: { confirmed: number; waitlisted: number },
): string {
  if (!rule.needSignup) return ''
  const confirmed = summary?.confirmed ?? 0
  const waitlisted = summary?.waitlisted ?? 0
  const max = rule.maxParticipants ?? 0

  if (max > 0) {
    const parts = [`已报 ${confirmed}/${max} 人`]
    if (waitlisted > 0) parts.push(`${waitlisted} 人候补`)
    return parts.join('，')
  }

  const parts: string[] = []
  if (confirmed > 0) parts.push(`${confirmed} 人已报名`)
  if (waitlisted > 0) parts.push(`${waitlisted} 人候补`)
  if (!parts.length) return '暂无人报名'
  return parts.join('，')
}

export function signupStatusText(status: number): string {
  if (status === SignupStatus.Confirmed) return '已报名'
  if (status === SignupStatus.Waitlisted) return '候补中'
  if (status === SignupStatus.Cancelled) return '已取消'
  return ''
}

export function formatUnixTime(ts: number): string {
  if (!ts) return ''
  return formatUnixDate(new Date(ts * 1000))
}

function formatPostSchedulePhase(
  startAt: number,
  endAt: number,
  nowSec = Math.floor(Date.now() / 1000),
): string {
  if (!startAt && !endAt) return ''
  if (endAt > 0 && nowSec > endAt) return '已结束'
  if (startAt > 0 && nowSec < startAt) return '未开始'
  if (startAt > 0 && (endAt <= 0 || nowSec <= endAt)) return '进行中'
  if (endAt > 0 && nowSec <= endAt) return '进行中'
  return ''
}

function formatPostScheduleCompact(startAt: number, endAt: number): string {
  if (!startAt && !endAt) return ''
  if (startAt && endAt) {
    const a = new Date(startAt * 1000)
    const b = new Date(endAt * 1000)
    const sameDay =
      a.getFullYear() === b.getFullYear() &&
      a.getMonth() === b.getMonth() &&
      a.getDate() === b.getDate()
    if (sameDay) {
      return `${formatUnixCompactDate(a)} ${pad(a.getHours())}:${pad(a.getMinutes())}–${pad(b.getHours())}:${pad(b.getMinutes())}`
    }
    return `${formatUnixCompact(startAt)}–${formatUnixCompact(endAt)}`
  }
  return formatUnixCompact(startAt || endAt)
}

function formatUnixCompactDate(d: Date): string {
  const now = new Date()
  if (d.getFullYear() === now.getFullYear()) {
    return `${d.getMonth() + 1}-${d.getDate()}`
  }
  return `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()}`
}

function formatUnixCompact(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  return `${formatUnixCompactDate(d)} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function formatIdlePriceBrief(rule?: { amount?: number; negotiable?: boolean }): string {
  if (!rule) return ''
  const price = Number(rule.amount) || 0
  if (price <= 0 && !rule.negotiable) return ''
  if (price <= 0) return '面议'
  return rule.negotiable ? `¥${price}·可议` : `¥${price}`
}

export type PostListPhaseTone = 'phase-pending' | 'phase-live' | 'phase-done'

export function formatPostListFactGroups(post: {
  category?: string
  startAt?: number
  endAt?: number
  extras?: {
    amount?: number
    negotiable?: boolean
    needSignup?: boolean
    maxParticipants?: number
    allowWaitlist?: boolean
    signupDeadlineAt?: number
    feeType?: number
  }
  signupSummary?: { confirmed: number; waitlisted: number }
}): {
  phase?: { text: string; tone: PostListPhaseTone }
  schedule?: string
  signup: string[]
  signupDeadline?: string
  fee?: string
} {
  const startAt = post.startAt ?? 0
  const endAt = post.endAt ?? 0
  const extras = post.extras

  let phase: { text: string; tone: PostListPhaseTone } | undefined
  const phaseText = formatPostSchedulePhase(startAt, endAt)
  if (phaseText === '未开始') phase = { text: phaseText, tone: 'phase-pending' }
  else if (phaseText === '进行中') phase = { text: phaseText, tone: 'phase-live' }
  else if (phaseText === '已结束') phase = { text: phaseText, tone: 'phase-done' }

  const schedule = formatPostScheduleCompact(startAt, endAt) || undefined

  const signup: string[] = []
  let signupDeadline: string | undefined
  if (extras?.needSignup) {
    const people = formatSignupBrief(extras, post.signupSummary)
    if (people) signup.push(people)
    const deadline = extras.signupDeadlineAt ?? 0
    if (deadline > 0) {
      signupDeadline = isSignupDeadlinePassed(deadline)
        ? '报名已截止'
        : `报名截止 ${formatUnixCompact(deadline)}`
      signup.push(signupDeadline)
    }
  }

  let fee: string | undefined
  if (post.category === 'idle') {
    fee = formatIdlePriceBrief(extras) || undefined
  }
  if (!fee && extras && hasPostFee(extras) && (Number(extras.amount) || 0) > 0) {
    fee = formatPostFeeBrief(extras) || undefined
  }

  return { phase, schedule, signup, signupDeadline, fee }
}

export function isSignupDeadlinePassed(
  deadline: number,
  nowSec = Math.floor(Date.now() / 1000),
): boolean {
  if (!deadline) return false
  return nowSec > deadline + 59
}

export function isEventEnded(endAt: number, nowSec = Math.floor(Date.now() / 1000)): boolean {
  return endAt > 0 && nowSec > endAt
}

export function canDiscussPost(post: { status?: number; endAt?: number }): boolean {
  return (post.status ?? 0) === PostStatus.Published && !isEventEnded(post.endAt ?? 0)
}

export function isSignupCheckInWindowOpen(
  startAt: number,
  endAt: number,
  nowSec = Math.floor(Date.now() / 1000),
): boolean {
  if (startAt > 0 && nowSec < startAt) return false
  if (endAt > 0 && nowSec > endAt) return false
  return true
}

export function hasCheckInTime(value?: string | number | null): boolean {
  if (value === undefined || value === null || value === '' || value === 0 || value === '0') {
    return false
  }
  return true
}

export function signupCheckInText(checkedInAt?: string | number | null): string {
  return hasCheckInTime(checkedInAt) ? '已核验' : '未核验'
}

export function signupListStatusText(item: {
  signupStatus: number
  signupStatusText?: string
}): string {
  if (item.signupStatusText) return item.signupStatusText
  if (item.signupStatus === SignupStatus.Confirmed) return '已报名'
  if (item.signupStatus === SignupStatus.Waitlisted) return '候补'
  if (item.signupStatus === SignupStatus.Cancelled) return '已取消'
  return ''
}

export function postStatusText(status: number): string {
  if (status === PostStatus.Pending) return '待审核'
  if (status === PostStatus.Published) return '已发布'
  if (status === PostStatus.Rejected) return '已驳回'
  if (status === PostStatus.Disabled) return '已下架'
  return '未知'
}

export function joinedListCornerBadge(post: {
  status?: number
  startAt?: number
  endAt?: number
  extras?: { needCheckIn?: boolean; needSignup?: boolean }
  mySignupStatus?: number
  myCheckedInAt?: number
}): { text: string; tone: string } | null {
  const status = post.status ?? 0
  if (status === PostStatus.Disabled) {
    return { text: '已下架', tone: 'disabled' }
  }
  if (status === PostStatus.Pending) {
    return { text: '待审核', tone: 'pending' }
  }
  if (status === PostStatus.Rejected) {
    return { text: '已驳回', tone: 'rejected' }
  }

  const signupStatus = post.mySignupStatus ?? 0
  if (signupStatus === SignupStatus.Waitlisted) {
    return { text: '候补中', tone: 'waitlisted' }
  }

  const needCheckIn = !!post.extras?.needSignup && !!post.extras?.needCheckIn
  if (signupStatus === SignupStatus.Confirmed && needCheckIn) {
    if (hasCheckInTime(post.myCheckedInAt)) {
      return { text: '已核验', tone: 'checked-in' }
    }

    if (isEventEnded(post.endAt ?? 0)) {
      return { text: '未核验', tone: 'check-in-missed' }
    }
    if (isSignupCheckInWindowOpen(post.startAt ?? 0, post.endAt ?? 0)) {
      return { text: '待核验', tone: 'check-in-pending' }
    }
  }

  const phaseText = formatPostSchedulePhase(post.startAt ?? 0, post.endAt ?? 0)
  if (phaseText === '未开始') return { text: phaseText, tone: 'phase-pending' }
  if (phaseText === '进行中') return { text: phaseText, tone: 'phase-live' }
  if (phaseText === '已结束') return { text: phaseText, tone: 'phase-done' }

  if (signupStatus === SignupStatus.Confirmed) {
    return { text: '已报名', tone: 'joined' }
  }
  return null
}

export function canShowJoinedCheckInCode(post: {
  status?: number
  startAt?: number
  endAt?: number
  extras?: { needCheckIn?: boolean; needSignup?: boolean }
  mySignupStatus?: number
  myCheckedInAt?: number
}): boolean {
  if ((post.status ?? 0) !== PostStatus.Published) return false
  if (!post.extras?.needSignup || !post.extras?.needCheckIn) return false
  if ((post.mySignupStatus ?? 0) !== SignupStatus.Confirmed) return false
  if (hasCheckInTime(post.myCheckedInAt)) return true
  return isSignupCheckInWindowOpen(post.startAt ?? 0, post.endAt ?? 0)
}

export function canShowJoinedCheckInCodeIcon(post: {
  status?: number
  startAt?: number
  endAt?: number
  extras?: { needCheckIn?: boolean; needSignup?: boolean }
  mySignupStatus?: number
  myCheckedInAt?: number
}): boolean {
  return canShowJoinedCheckInCode(post) && !hasCheckInTime(post.myCheckedInAt)
}

export function canDeletePublishedPost(post: { status?: number }): boolean {
  const status = post.status ?? -1
  return (
    status === PostStatus.Pending ||
    status === PostStatus.Rejected ||
    status === PostStatus.Disabled
  )
}

export function canDeleteJoinedPost(post: {
  status?: number
  endAt?: number
  myCheckedInAt?: number
}): boolean {
  if ((post.status ?? 0) === PostStatus.Disabled) return true
  if (hasCheckInTime(post.myCheckedInAt)) return true
  return isEventEnded(post.endAt ?? 0)
}
