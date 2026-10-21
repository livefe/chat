import type { Router } from 'vue-router'

const MASKED_MOBILE = /^\d{3}\*{4}\d{4}$/

export function sanitizeNickname(raw?: string | null): string {
  const name = (raw || '').trim()
  if (!name) return ''
  if (name.length === 11 && /^\d{11}$/.test(name)) return ''
  if (MASKED_MOBILE.test(name)) return ''
  return name
}

export function peerDisplayName(peer: { peerNickname?: string | null }): string {
  return sanitizeNickname(peer.peerNickname)
}

export interface OpenUserProfileOptions {
  from?: string
  name?: string
  avatarUrl?: string
  postId?: number | string
  chatFrom?: string

  peerId?: number | string
  peerName?: string
}

export function openUserProfile(router: Router, userId: number, opts: OpenUserProfileOptions = {}) {
  if (!userId) return
  const query: Record<string, string> = {}
  if (opts.from) query.from = opts.from
  if (opts.name) query.name = opts.name
  if (opts.postId != null) query.postId = String(opts.postId)
  if (opts.chatFrom) query.chatFrom = opts.chatFrom
  if (opts.peerId != null && Number(opts.peerId) > 0) query.peerId = String(opts.peerId)
  if (opts.peerName) query.peerName = opts.peerName

  const previewName = (opts.name || '').trim()
  const previewAvatar = (opts.avatarUrl || '').trim()
  const hasPreview = !!(previewName || previewAvatar)

  router.push({
    name: 'UserProfile',
    params: { id: String(userId) },
    query,
    state: hasPreview
      ? { profilePreview: { name: previewName, avatarUrl: previewAvatar } }
      : undefined,
  })
}
