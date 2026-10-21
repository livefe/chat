import { defineStore } from 'pinia'

import { recallMessage as recallMessageApi } from '@/api/message'
import { POST_SUBSCRIBE_TIMEOUT_MS, ROOM_SUBSCRIBE_TIMEOUT_MS } from '@/constants/chat'
import type { WSMessage } from '@/utils/chat/chatWs'

import { useUserStore } from '../user'
import { createActiveChannelHelpers } from './activeChannel'
import { createConversationPreviewHelpers } from './conversationPreview'
import { createIncomingHandler } from './incoming'
import { loadMoreActiveHistory } from './loadHistory'
import { createOptimisticHelpers } from './messages'
import { createNearbyHelpers } from './nearby'
import { createPendingSubscribeWait } from './pendingSubscribe'
import { createPrivateClearedHelpers } from './privateCleared'
import { createPrivateChatHelpers } from './privateChatSession'
import { catchUpAfterReconnect } from './reconnect'
import { createRoomSessionHelpers } from './roomSession'
import { createScopedFeedHelpers } from './scopedFeed'
import { createSendHelpers } from './send'
import { clearChatSessionState } from './sessionReset'
import { createChatState } from './state'
import { createSubscriptionHelpers } from './subscriptions'
import type { ChatLocationSource, ChatNotice } from './types'
import { createUnreadHelpers } from './unread'
import { createWsSession, type WsSessionDeps } from './wsSession'

export type { ChatNotice, SendFailReason, SendResult } from './types'

export const useChatStore = defineStore('chat', () => {
  const userStore = useUserStore()
  const state = createChatState()
  const { activeChannel, currentRoomId, currentPeerId, rooms, nearbyOnlineCount, messagesLoading } =
    state

  let locationSource: ChatLocationSource = () => null
  let supportEventHandler: (env: WSMessage) => void = () => {}
  const noticeHandlers = new Set<(notice: ChatNotice) => void>()

  const getLocation: ChatLocationSource = () => locationSource()

  function setLocationSource(source: ChatLocationSource) {
    locationSource = source
  }

  function setSupportEventHandler(handler: (env: WSMessage) => void) {
    supportEventHandler = handler
  }

  function onNotice(handler: (notice: ChatNotice) => void): () => void {
    noticeHandlers.add(handler)
    return () => noticeHandlers.delete(handler)
  }

  function notify(notice: ChatNotice) {
    for (const handler of noticeHandlers) handler(notice)
  }

  const { setPrivateClearedBefore, isClearedMessage, clearAllPrivateCleared } =
    createPrivateClearedHelpers()

  const {
    getUnread,
    markPeerRead,
    flushPeerRead,
    acknowledgePeerRead,
    incrPeerUnread,
    applyConversationUnread,
    clearAllUnread,
    clearAllReadAckTimers,
    refreshUnreadSummary,
    tabUnreadBadge,
  } = createUnreadHelpers({ notify })

  const roomSubscribeWait = createPendingSubscribeWait({
    timeoutMs: ROOM_SUBSCRIBE_TIMEOUT_MS,
    timeoutMessage: '加入聊天区超时，请重试',
  })
  const postSubscribeWait = createPendingSubscribeWait({
    timeoutMs: POST_SUBSCRIBE_TIMEOUT_MS,
    timeoutMessage: '加入讨论超时，请重试',
  })

  const { upsertConversationPreview, clearConversationPreview } =
    createConversationPreviewHelpers(state)

  const {
    applyAckEverywhere,
    failOptimisticEverywhere,
    failAllOptimisticEverywhere,
    insertOptimistic,
    insertOptimisticPost,
    buildOptimistic,
  } = createOptimisticHelpers({
    ...state,
    getSelfProfile: () => userStore.info,
    upsertConversationPreview,
  })

  const {
    getPrivateChatMessages,
    refreshPrivateChatMessages,
    enterPrivateChat,
    leavePrivateChat,
    deleteConversation,
  } = createPrivateChatHelpers({
    ...state,
    setPrivateClearedBefore,
    markPeerRead,
    flushPeerRead,
    clearConversationPreview,
    refreshUnreadSummary,
  })

  const {
    currentMessages,
    currentHistoryState,
    setCurrentHistoryState,
    currentMessageList,
    historyLoading,
    historyHasMore,
  } = createActiveChannelHelpers(state)

  const {
    setRoomOnlineCount,
    onlineCountForRoom,
    setPostOnlineCount,
    onlineCountForPost,
    getPostMessages,
    getPostHistory,
    refreshRoomMessages,
    refreshPostMessages,
    applyRoomMessages,
    loadMorePostMessages,
  } = createScopedFeedHelpers({ ...state, getLocation })

  const sessionDeps: WsSessionDeps = {
    connected: state.connected,
    getToken: () => userStore.accessToken,
    tryRefresh: () => userStore.tryRefresh(),
    reportLocation: () => {},
    onMessage: () => {},
    onOpen: () => {},
    onBeforeReconnect: () => {},
    onDisconnectCleanup: () => {},
  }
  const ws = createWsSession(sessionDeps)
  const { connect, ensureAlive, waitForOpen, disconnect } = ws

  function reportLocation() {
    const loc = getLocation()
    if (!loc) return
    ws.getClient()?.send({ type: 'location', lat: loc.lat, lng: loc.lng })
  }
  sessionDeps.reportLocation = reportLocation

  const {
    clearPendingSend,
    clearAllPendingSends,
    send,
    sendToNearby,
    sendToRoom,
    sendToPrivateChat,
    sendToPost,
  } = createSendHelpers({
    ...state,
    getLocation,
    reportLocation,
    getClient: ws.getClient,
    connect,
    waitForOpen,
    insertOptimistic,
    insertOptimisticPost,
    buildOptimistic,
    failOptimisticEverywhere,
    notify,
  })

  const { handleIncoming, applyRecallEverywhere } = createIncomingHandler({
    ...state,
    getSelfId: () => userStore.info?.id,
    clearPendingSend,
    applyAckEverywhere,
    failOptimisticEverywhere,
    roomSubscribeMatches: roomSubscribeWait.matches,
    rejectRoomSubscribe: roomSubscribeWait.reject,
    resolveRoomSubscribe: roomSubscribeWait.resolve,
    postSubscribeMatches: postSubscribeWait.matches,
    rejectPostSubscribe: postSubscribeWait.reject,
    resolvePostSubscribe: postSubscribeWait.resolve,
    setRoomOnlineCount,
    setPostOnlineCount,
    isClearedMessage,
    upsertConversationPreview,
    acknowledgePeerRead,
    incrPeerUnread,
    notify,
    forwardSupportEvent: (env) => supportEventHandler(env),
  })
  sessionDeps.onMessage = handleIncoming

  const { refreshNearby, clearNearbyMessages } = createNearbyHelpers({
    ...state,
    getLocation,
    reportLocation,
  })

  const subs = createSubscriptionHelpers({
    ...state,
    getClient: ws.getClient,
    getActiveLocation: getLocation,
  })

  function clearSessionState() {
    clearAllUnread()
    clearAllPrivateCleared()
    clearChatSessionState(state)
  }

  function abortPendingWork() {
    roomSubscribeWait.rejectAll(new Error('连接已断开'))
    postSubscribeWait.rejectAll(new Error('连接已断开'))
    clearAllPendingSends()
    failAllOptimisticEverywhere()
  }

  sessionDeps.onOpen = () => {
    reportLocation()
    subs.resyncRoomSubscriptions()
    subs.resyncPostSubscriptions()
    void refreshUnreadSummary()
    void catchUpAfterReconnect({
      activeChannel,
      currentRoomId,
      currentPeerId,
      hasLocation: () => getLocation() !== null,
      refreshNearby,
      refreshRoomMessages,
      refreshPrivateChatMessages,
      refreshPostMessages,
      feedPostIds: subs.feedPostIds,
      notify,
    })
  }
  sessionDeps.onBeforeReconnect = abortPendingWork
  sessionDeps.onDisconnectCleanup = (opts) => {
    clearAllReadAckTimers()
    abortPendingWork()
    if (opts?.clearSession) clearSessionState()
  }

  const { roomName, enterRoom, leaveRoom, subscribePostFeed, unsubscribePostFeed } =
    createRoomSessionHelpers({
      ...state,
      connect,
      getLocation,
      waitRoomSubscribe: roomSubscribeWait.wait,
      rejectRoomSubscribe: roomSubscribeWait.reject,
      acquireRoomSubscription: subs.acquireRoomSubscription,
      releaseRoomSubscription: subs.releaseRoomSubscription,
      sendRoomSubscribe: subs.sendRoomSubscribe,
      applyRoomMessages,
      setRoomOnlineCount,
      waitPostSubscribe: postSubscribeWait.wait,
      postSubscribeMatches: postSubscribeWait.matches,
      rejectPostSubscribe: postSubscribeWait.reject,
      acquirePostSubscription: subs.acquirePostSubscription,
      releasePostSubscription: subs.releasePostSubscription,
      sendPostSubscribe: subs.sendPostSubscribe,
    })

  async function loadMoreHistory(): Promise<boolean> {
    return loadMoreActiveHistory({
      activeChannel,
      currentRoomId,
      currentPeerId,
      getLocation,
      getState: currentHistoryState,
      setState: setCurrentHistoryState,
      list: currentMessageList(),
      setPrivateClearedBefore,
    })
  }

  async function recallMessage(id: number) {
    if (id <= 0) return
    await recallMessageApi(id)
    applyRecallEverywhere(id)
  }

  return {
    activeChannel,
    rooms,
    nearbyOnlineCount,
    currentMessages,
    historyLoading,
    historyHasMore,
    messagesLoading,
    tabUnreadBadge,
    getUnread,
    applyConversationUnread,
    conversationPreviews: state.conversationPreviews,
    refreshUnreadSummary,
    roomName,
    connect,
    disconnect,
    clearSessionState,
    ensureAlive,
    send,
    sendToNearby,
    sendToRoom,
    sendToPrivateChat,
    sendToPost,
    recallMessage,
    getPostMessages,
    getPostHistory,
    refreshPostMessages,
    subscribePostFeed,
    unsubscribePostFeed,
    loadMorePostMessages,
    refreshNearby,
    clearNearbyMessages,
    enterRoom,
    leaveRoom,
    onlineCountForRoom,
    onlineCountForPost,
    enterPrivateChat,
    leavePrivateChat,
    getPrivateChatMessages,
    deleteConversation,
    loadMoreHistory,
    setLocationSource,
    setSupportEventHandler,
    onNotice,
  }
})
