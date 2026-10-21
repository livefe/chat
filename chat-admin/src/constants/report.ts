export const ReportStatus = {
  Pending: 0 as const,
  Resolved: 1 as const,
  Rejected: 2 as const,
}

export const ReportTargetType = {
  User: 1,
  Message: 2,
  Post: 3,
} as const

export const REPORT_TARGET_TYPE_OPTIONS = [
  { label: '用户', value: ReportTargetType.User },
  { label: '消息', value: ReportTargetType.Message },
  { label: '帖子', value: ReportTargetType.Post },
]

export const REPORT_REASON_OPTIONS = [
  { label: '垃圾广告', value: 1 },
  { label: '色情低俗', value: 2 },
  { label: '辱骂骚扰', value: 3 },
  { label: '违法违规', value: 4 },
  { label: '虚假信息', value: 5 },
  { label: '侵权', value: 6 },
  { label: '其他', value: 7 },
]

export const REPORT_STATUS_OPTIONS = [
  { label: '待处理', value: 0 },
  { label: '已处理', value: 1 },
  { label: '已驳回', value: 2 },
]

export const REPORT_HANDLE_ACTION_OPTIONS = [
  { label: '无操作', value: 0 },
  { label: '警告', value: 1 },
  { label: '删除内容', value: 2 },
  { label: '禁言用户', value: 3 },
  { label: '禁用账号', value: 4 },
  { label: '下架帖子', value: 5 },
]
