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
}

export type ImMineMode = 'staff' | 'support'
