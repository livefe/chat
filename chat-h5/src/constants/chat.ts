export const MessageScope = {
  Nearby: 1,
  Room: 2,
  Private: 3,
  Post: 4,

  SystemHint: 90,
} as const

export type MessageScope = (typeof MessageScope)[keyof typeof MessageScope]

export type ChatChannel = 'nearby' | 'room' | 'private' | 'post'

export type ActiveChatChannel = Exclude<ChatChannel, 'post'>

export type WsScope = ChatChannel | 'support'

const CHANNEL_TO_SCOPE: Record<ChatChannel, MessageScope> = {
  nearby: MessageScope.Nearby,
  room: MessageScope.Room,
  private: MessageScope.Private,
  post: MessageScope.Post,
}

export function channelToMessageScope(channel: ChatChannel): MessageScope {
  return CHANNEL_TO_SCOPE[channel]
}

export function messageScopeToChannel(scope: number): ChatChannel | null {
  switch (scope) {
    case MessageScope.Nearby:
      return 'nearby'
    case MessageScope.Room:
      return 'room'
    case MessageScope.Private:
      return 'private'
    case MessageScope.Post:
      return 'post'
    default:
      return null
  }
}

export const MessageContentTypes = {
  Text: 1,
  Image: 2,
  Voice: 3,
  Video: 4,
  File: 5,
  Location: 6,
  Post: 7,
} as const

export const RECALL_WINDOW_SEC = 120

export const CHAT_PAGE_LIMIT = 50
export const ROOM_MARQUEE_LIMIT = 5

export const PENDING_SEND_TIMEOUT_MS = 15_000
export const ROOM_SUBSCRIBE_TIMEOUT_MS = 10_000
export const POST_SUBSCRIBE_TIMEOUT_MS = 10_000
export const READ_ACK_DEBOUNCE_MS = 300
export const RECONNECT_CATCHUP_POST_LIMIT = 1
