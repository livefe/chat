import type { Ref } from 'vue'

import { RECONNECT_CATCHUP_POST_LIMIT, type ActiveChatChannel } from '@/constants/chat'

import type { ChatNotice } from './types'

export async function catchUpAfterReconnect(opts: {
  activeChannel: Ref<ActiveChatChannel>
  currentRoomId: Ref<number>
  currentPeerId: Ref<number>
  hasLocation: () => boolean
  refreshNearby: (opts: {
    includeMessages: boolean
    replaceMessages: boolean
    syncServerLocation: boolean
  }) => Promise<unknown>
  refreshRoomMessages: (roomId: number) => Promise<unknown>
  refreshPrivateChatMessages: (peerId: number) => Promise<unknown>
  refreshPostMessages: (postId: number) => Promise<unknown>
  feedPostIds: () => number[]
  notify: (notice: ChatNotice) => void
}): Promise<void> {
  const tasks: Promise<unknown>[] = []

  if (opts.activeChannel.value === 'nearby' && opts.hasLocation()) {
    tasks.push(
      opts.refreshNearby({
        includeMessages: true,
        replaceMessages: false,
        syncServerLocation: false,
      }),
    )
  } else if (opts.activeChannel.value === 'room' && opts.currentRoomId.value > 0) {
    tasks.push(opts.refreshRoomMessages(opts.currentRoomId.value))
  } else if (opts.activeChannel.value === 'private' && opts.currentPeerId.value > 0) {
    tasks.push(opts.refreshPrivateChatMessages(opts.currentPeerId.value))
  }

  for (const postId of opts.feedPostIds().slice(0, RECONNECT_CATCHUP_POST_LIMIT)) {
    tasks.push(opts.refreshPostMessages(postId))
  }

  if (tasks.length === 0) return
  const results = await Promise.allSettled(tasks)
  const firstRejected = results.find((r) => r.status === 'rejected')
  if (firstRejected?.status === 'rejected') {
    opts.notify({ type: 'syncFailed', error: firstRejected.reason })
  }
}
