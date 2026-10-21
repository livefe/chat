export const MessageScope = {
  Nearby: 1,
  Room: 2,
  Private: 3,
  Post: 4,
} as const

export const MessageContentType = {
  Text: 1,
  Image: 2,
  Voice: 3,
  Video: 4,
  File: 5,
  Location: 6,
  Post: 7,
} as const

export const MESSAGE_SCOPE_OPTIONS = [
  { label: '附近', value: MessageScope.Nearby },
  { label: '聊天区', value: MessageScope.Room },
  { label: '私聊', value: MessageScope.Private },
  { label: '帖子讨论', value: MessageScope.Post },
]
