import type { Ref } from 'vue'

import { postMessages, roomMessages, type MessageItem } from '@/api/message'
import { CHAT_PAGE_LIMIT } from '@/constants/chat'
import { mergeMessageLists, normalizeMessages } from '@/utils/chat/chatMessages'

import { loadOlderMessages } from './history'
import { emptyHistory, historyFromBatch, type ChatLocationSource, type HistoryState } from './types'

function setOnlineCount(onlineById: Ref<Record<number, number>>, id: number, onlineCount: number) {
  if (id <= 0 || onlineCount < 0) return
  if (onlineById.value[id] !== onlineCount) {
    onlineById.value = { ...onlineById.value, [id]: onlineCount }
  }
}

function getOnlineCount(onlineById: Ref<Record<number, number>>, id: number): number {
  if (id <= 0) return 0
  return onlineById.value[id] ?? 0
}

function createIdKeyedFeed(opts: {
  msgsMap: Ref<Map<number, MessageItem[]>>
  history: Ref<Record<number, HistoryState>>
  fetchPage: (id: number, beforeId: number) => Promise<MessageItem[]>
}) {
  const generations = new Map<number, number>()

  function nextGen(id: number): number {
    const gen = (generations.get(id) ?? 0) + 1
    generations.set(id, gen)
    return gen
  }

  function isCurrentGen(id: number, gen: number): boolean {
    return generations.get(id) === gen
  }

  function getMessages(id: number) {
    return opts.msgsMap.value.get(id) || []
  }

  function getHistory(id: number): HistoryState {
    return opts.history.value[id] ?? emptyHistory()
  }

  function replaceList(id: number, next: MessageItem[]) {
    opts.msgsMap.value.set(id, next)
  }

  async function refreshMessages(id: number) {
    if (id <= 0) return
    const gen = nextGen(id)
    const batch = await opts.fetchPage(id, 0)
    if (!isCurrentGen(id, gen)) return
    applyBatch(id, batch)
  }

  function applyBatch(id: number, batch: MessageItem[]) {
    if (id <= 0) return
    const existing = opts.msgsMap.value.get(id)

    if (!existing || existing.length === 0) {
      replaceList(id, batch)
      opts.history.value = { ...opts.history.value, [id]: historyFromBatch(batch) }
      return
    }
    if (batch.length === 0) return

    const existingNewestRealId = [...existing].reverse().find((m) => m.id > 0)?.id ?? 0
    const batchOldestId = batch[0]?.id ?? 0

    if (existingNewestRealId === 0 || existingNewestRealId >= batchOldestId) {
      replaceList(id, mergeMessageLists(existing, batch))
      return
    }

    const optimisticOnly = existing.filter((m) => m.id <= 0)
    const next = optimisticOnly.length ? normalizeMessages([...batch, ...optimisticOnly]) : batch
    replaceList(id, next)
    opts.history.value = { ...opts.history.value, [id]: { hasMore: true, loading: false } }
  }

  async function loadMoreMessages(id: number): Promise<boolean> {
    if (id <= 0) return false
    return loadOlderMessages({
      getState: () => getHistory(id),
      setState: (state) => {
        opts.history.value = { ...opts.history.value, [id]: state }
      },
      list: opts.msgsMap.value.get(id) ?? null,
      fetch: async (beforeId) => opts.fetchPage(id, beforeId),
    })
  }

  return {
    getMessages,
    getHistory,
    refreshMessages,
    applyMessages: applyBatch,
    loadMoreMessages,
  }
}

export interface ScopedFeedDeps {
  roomMsgsMap: Ref<Map<number, MessageItem[]>>
  postMsgsMap: Ref<Map<number, MessageItem[]>>
  roomHistory: Ref<Record<number, HistoryState>>
  postHistory: Ref<Record<number, HistoryState>>
  roomOnlineById: Ref<Record<number, number>>
  postOnlineById: Ref<Record<number, number>>
  getLocation: ChatLocationSource
}

export function createScopedFeedHelpers(deps: ScopedFeedDeps) {
  const roomFeed = createIdKeyedFeed({
    msgsMap: deps.roomMsgsMap,
    history: deps.roomHistory,
    fetchPage: async (roomId, beforeId) => {
      const loc = deps.getLocation()
      const res = await roomMessages(roomId, beforeId, CHAT_PAGE_LIMIT, loc?.lat, loc?.lng)
      return normalizeMessages(res.data.list)
    },
  })

  const postFeed = createIdKeyedFeed({
    msgsMap: deps.postMsgsMap,
    history: deps.postHistory,
    fetchPage: async (postId, beforeId) => {
      const res = await postMessages(postId, beforeId, CHAT_PAGE_LIMIT)
      return normalizeMessages(res.data.list)
    },
  })

  return {
    setRoomOnlineCount: (id: number, n: number) => setOnlineCount(deps.roomOnlineById, id, n),
    onlineCountForRoom: (id: number) => getOnlineCount(deps.roomOnlineById, id),
    setPostOnlineCount: (id: number, n: number) => setOnlineCount(deps.postOnlineById, id, n),
    onlineCountForPost: (id: number) => getOnlineCount(deps.postOnlineById, id),
    getPostMessages: postFeed.getMessages,
    getPostHistory: postFeed.getHistory,
    refreshRoomMessages: roomFeed.refreshMessages,
    refreshPostMessages: postFeed.refreshMessages,
    applyRoomMessages: roomFeed.applyMessages,
    loadMorePostMessages: postFeed.loadMoreMessages,
  }
}
