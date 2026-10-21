import { ElMessage, ElMessageBox } from 'element-plus'
import { MessageContentType } from '@/constants/message'
import {
  formatMessageLocation,
  formatMessagePost,
  formatMessageVoice,
  parseMessageMediaUrl,
} from '@/utils/messageDisplay'

export function imFileName(content: string) {
  try {
    const body = JSON.parse(content) as { name?: string }
    if (body.name) return body.name
  } catch {}
  return '文件'
}

function imTruncateName(name: string, maxLen = 28) {
  const n = (name || '').trim() || '文件'
  if (n.length <= maxLen) return n
  const dot = n.lastIndexOf('.')
  const ext = dot > 0 && n.length - dot <= 9 ? n.slice(dot) : ''
  const budget = Math.max(8, maxLen - ext.length - 1)
  return `${n.slice(0, budget)}…${ext}`
}

export function imMessagePreview(m: { contentType: number; content: string; recalled?: boolean }) {
  if (m.recalled) return '[已撤回]'
  if (m.contentType === MessageContentType.Image) return '[图片]'
  if (m.contentType === MessageContentType.Video) return '[视频]'
  if (m.contentType === MessageContentType.Voice) return formatMessageVoice(m.content)
  if (m.contentType === MessageContentType.Location) return formatMessageLocation(m.content)
  if (m.contentType === MessageContentType.Post) return formatMessagePost(m.content)
  if (m.contentType === MessageContentType.File) {
    const name = imFileName(m.content)
    return name === '文件' ? '[文件]' : `[文件] ${imTruncateName(name)}`
  }
  return (m.content || '').trim()
}

export function imFormatTime(sec?: number) {
  if (!sec) return ''
  const d = new Date(sec * 1000)
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  const hm = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  if (sameDay) return hm
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  const isYesterday =
    d.getFullYear() === yesterday.getFullYear() &&
    d.getMonth() === yesterday.getMonth() &&
    d.getDate() === yesterday.getDate()
  if (isYesterday) return `昨天 ${hm}`
  if (d.getFullYear() === now.getFullYear()) {
    return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${hm}`
  }
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export function imFormatWait(sec?: number) {
  const s = Math.max(0, sec || 0)
  if (s < 60) return `${s}s`
  if (s < 3600) return `${Math.floor(s / 60)}分${s % 60}秒`
  return `${Math.floor(s / 3600)}小时${Math.floor((s % 3600) / 60)}分`
}

export function imWaitTone(sec?: number): 'ok' | 'warn' | 'danger' {
  const s = Math.max(0, sec || 0)
  if (s >= 300) return 'danger'
  if (s >= 120) return 'warn'
  return 'ok'
}

export function imInitials(name?: string) {
  const n = (name || '').trim()
  return n ? n.slice(0, 1) : '?'
}

export async function imConfirmDownload(content: string) {
  const url = parseMessageMediaUrl(content)
  const name = imFileName(content)
  if (!url) {
    ElMessage.error('文件地址无效')
    return
  }
  try {
    await ElMessageBox.confirm(`确定下载「${imTruncateName(name)}」？`, '下载文件', {
      confirmButtonText: '下载',
      cancelButtonText: '取消',
      type: 'info',
    })
  } catch {
    return
  }
  const a = document.createElement('a')
  a.href = url
  a.download = name
  a.target = '_blank'
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  a.remove()
}
