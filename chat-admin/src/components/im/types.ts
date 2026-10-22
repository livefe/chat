export interface ImThreadMessage {
  id: number
  conversationId?: number
  senderKind?: number
  senderId: number
  senderName?: string
  senderAvatarUrl?: string
  contentType: number
  content: string
  createdAt: number
  recalled?: boolean
  clientMsgId?: string
  /** 本地发送状态：未收到服务端确认前为 sending，超时或失败为 failed */
  sendState?: 'sending' | 'failed'
}

export type ImMineMode = 'staff' | 'support'

export interface ImQuickReply {
  title: string
  content: string
}
