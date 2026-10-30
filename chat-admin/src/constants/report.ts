import { FILTER_ALL } from './index'

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

export const ReportHandleAction = {
  None: 0,
  Warn: 1,
  DeleteContent: 2,
  Mute: 3,
  DisableUser: 4,
  DisablePost: 5,
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
  { label: '待处理', value: ReportStatus.Pending },
  { label: '已处理', value: ReportStatus.Resolved },
  { label: '已驳回', value: ReportStatus.Rejected },
]

export const REPORT_STATUS_FILTER_OPTIONS = [
  ...REPORT_STATUS_OPTIONS,
  { label: '全部', value: FILTER_ALL },
]

export const REPORT_HANDLE_ACTION_OPTIONS = [
  { label: '无操作', value: ReportHandleAction.None },
  { label: '警告（发送站内通知）', value: ReportHandleAction.Warn },
  { label: '删除消息', value: ReportHandleAction.DeleteContent },
  { label: '禁言用户', value: ReportHandleAction.Mute },
  { label: '禁用账号', value: ReportHandleAction.DisableUser },
  { label: '下架帖子', value: ReportHandleAction.DisablePost },
]
