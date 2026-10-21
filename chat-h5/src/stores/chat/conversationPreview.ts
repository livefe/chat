import type { Ref } from 'vue'

import type { ConversationItem } from '@/api/conversation'
import { sanitizeNickname } from '@/utils/user'

export function createConversationPreviewHelpers(deps: {
  conversationPreviews: Ref<Record<number, ConversationItem>>
}) {
  function upsertConversationPreview(opts: {
    peerId: number
    lastContent: string
    lastActiveAt: number
    peerNickname?: string
    peerAvatar?: string
  }) {
    if (opts.peerId <= 0) return
    const existing = deps.conversationPreviews.value[opts.peerId]
    deps.conversationPreviews.value = {
      ...deps.conversationPreviews.value,
      [opts.peerId]: {
        peerId: opts.peerId,
        peerNickname:
          sanitizeNickname(opts.peerNickname) || sanitizeNickname(existing?.peerNickname) || '',
        peerAvatar: opts.peerAvatar || existing?.peerAvatar || '',
        peerMobile: existing?.peerMobile || '',
        lastContent: opts.lastContent,
        lastActiveAt: opts.lastActiveAt || existing?.lastActiveAt || 0,
        unreadCount: existing?.unreadCount ?? 0,
      },
    }
  }

  function clearConversationPreview(peerId: number) {
    if (peerId <= 0 || !deps.conversationPreviews.value[peerId]) return
    const next = { ...deps.conversationPreviews.value }
    delete next[peerId]
    deps.conversationPreviews.value = next
  }

  return { upsertConversationPreview, clearConversationPreview }
}
