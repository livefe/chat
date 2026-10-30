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
  sendState?: 'sending' | 'failed'
}

export type ImMineMode = 'staff' | 'support'

export interface ImQuickReply {
  title: string
  content: string
}
