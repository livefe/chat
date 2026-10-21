import { showConfirmDialog } from 'vant'
import { useRouter, type Router } from 'vue-router'

import { ReportTargetMessage, ReportTargetPost, ReportTargetUser } from '@/api/report'
import type { MessageItem } from '@/api/message'
import { messagePreview } from '@/utils/chat/chatMessageContent'
import type { NavMoreAction } from '@/types/nav'

export const ACTION_REPORT_USER = '举报'
export const ACTION_REPORT_CONTENT = '举报该内容'

export const reportUserAction: NavMoreAction = {
  name: ACTION_REPORT_USER,
  color: '#d4544a',
}

export const reportContentAction: NavMoreAction = {
  name: ACTION_REPORT_CONTENT,
  color: '#d4544a',
}

function pushReport(
  router: Router,
  targetType: number,
  targetId: number,
  preview?: { title?: string; subtitle?: string; avatar?: string },
) {
  router.push({
    path: '/report',
    query: {
      targetType: String(targetType),
      targetId: String(targetId),
      title: preview?.title || '',
      subtitle: preview?.subtitle || '',
      avatar: preview?.avatar || '',
    },
  })
}

export function openUserReportSheet(
  router: ReturnType<typeof useRouter>,
  userId: number,
  nickname: string,
  avatar?: string,
) {
  showConfirmDialog({
    title: '举报用户',
    message: `确定要举报「${nickname}」吗？`,
  })
    .then(() => {
      pushReport(router, ReportTargetUser, userId, {
        title: nickname,
        avatar,
      })
    })
    .catch(() => {})
}

export function openPostReportSheet(
  router: ReturnType<typeof useRouter>,
  postId: number,
  title: string,
  content?: string,
) {
  showConfirmDialog({
    title: '举报内容',
    message: `确定要举报「${title}」吗？`,
  })
    .then(() => {
      pushReport(router, ReportTargetPost, postId, {
        title,
        subtitle: content,
      })
    })
    .catch(() => {})
}

export function openMessageReportSheet(router: ReturnType<typeof useRouter>, message: MessageItem) {
  const messageId = message.id
  if (!messageId) return

  const previewRaw = messagePreview(message.contentType, message.content)
  const preview = previewRaw.length > 40 ? `${previewRaw.slice(0, 40)}…` : previewRaw

  showConfirmDialog({
    title: '举报消息',
    message: `确定要举报这条消息吗？\n${preview}`,
  })
    .then(() => {
      pushReport(router, ReportTargetMessage, messageId, {
        title: message.fromName || '消息',
        subtitle: previewRaw,
        avatar: message.fromAvatar,
      })
    })
    .catch(() => {})
}
