import { FILTER_ALL } from './index'

export const FeedbackStatus = {
  Pending: 0 as const,
  Resolved: 1 as const,
  Closed: 2 as const,
}

export const FEEDBACK_TYPE_OPTIONS = [
  { label: '产品建议', value: 1 },
  { label: '功能异常', value: 2 },
  { label: '体验问题', value: 3 },
  { label: '账号问题', value: 4 },
  { label: '其他', value: 5 },
]

export const FEEDBACK_STATUS_OPTIONS = [
  { label: '待处理', value: FeedbackStatus.Pending },
  { label: '已解决', value: FeedbackStatus.Resolved },
  { label: '已关闭', value: FeedbackStatus.Closed },
]

export const FEEDBACK_STATUS_FILTER_OPTIONS = [
  ...FEEDBACK_STATUS_OPTIONS,
  { label: '全部', value: FILTER_ALL },
]

export function feedbackStatusTagType(status: number): 'success' | 'info' | 'danger' {
  if (status === FeedbackStatus.Resolved) return 'success'
  if (status === FeedbackStatus.Closed) return 'info'
  return 'danger'
}
