import type { ReportItem } from '@/api/report'
import { MESSAGE_SCOPE_OPTIONS, MessageContentType } from '@/constants/message'
import { ReportStatus, ReportTargetType } from '@/constants/report'
import { formatMessageSummary } from '@/utils/messageDisplay'

export function reportStatusTagType(status: number): 'success' | 'info' | 'danger' {
  if (status === ReportStatus.Resolved) return 'success'
  if (status === ReportStatus.Rejected) return 'info'
  return 'danger'
}

export function messageScopeLabel(scope?: number) {
  return MESSAGE_SCOPE_OPTIONS.find((o) => o.value === scope)?.label ?? '消息'
}

export function reportTargetPreview(row: ReportItem, maxLen = 60) {
  const ctx = row.context
  if (row.targetType === ReportTargetType.Post) {
    return ctx.postTitle || `帖子 #${row.targetId}`
  }
  if (row.targetType === ReportTargetType.Message) {
    const text = formatMessageSummary(
      ctx.messageType ?? MessageContentType.Text,
      ctx.messageContent ?? '',
    )
    if (!text) return `消息 #${row.targetId}`
    return text.length > maxLen ? `${text.slice(0, maxLen)}…` : text
  }
  return ctx.nickname || row.targetUserName || `用户 #${row.targetId}`
}
