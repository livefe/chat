import type { MessageItem } from '@/api/message'
import { prependMessages } from '@/utils/chat/chatMessages'

import { emptyHistory, historyFromBatch, type HistoryState } from './types'

export async function loadOlderMessages(opts: {
  getState: () => HistoryState
  setState: (state: HistoryState) => void
  list: MessageItem[] | null
  fetch: (beforeId: number) => Promise<MessageItem[] | null>
}): Promise<boolean> {
  const state = opts.getState()
  if (state.loading || !state.hasMore) return false

  const list = opts.list
  if (!list || list.length === 0) {
    opts.setState(emptyHistory())
    return false
  }

  const oldest = list[0]
  if (!oldest?.id) {
    opts.setState({ ...state, hasMore: false })
    return false
  }

  opts.setState({ ...state, loading: true })
  try {
    const batch = await opts.fetch(oldest.id)
    if (batch === null) {
      opts.setState({ ...state, loading: false })
      return false
    }
    prependMessages(list, batch)
    opts.setState(historyFromBatch(batch))
    return batch.length > 0
  } catch (e) {
    opts.setState({ ...state, loading: false })
    throw e
  }
}
