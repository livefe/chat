export const SupportSessionStatus = {
  Active: 1,
  Closed: 2,
  Queued: 3,
} as const

export const SUPPORT_CLOSE_REASON_OPTIONS = [
  { value: 'resolved', label: '已解决' },
  { value: 'no_response', label: '无应答' },
  { value: 'duplicate', label: '重复' },
  { value: 'other', label: '其他' },
]

/** 排队超过该秒数视为久等 */
export const SUPPORT_LONG_WAIT_SEC = 120

export const SUPPORT_QUICK_REPLY_MAX = 50
