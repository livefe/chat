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
  { label: '待处理', value: 0 },
  { label: '已解决', value: 1 },
  { label: '已关闭', value: 2 },
]
