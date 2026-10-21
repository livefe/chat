import type { MessageItem } from '@/api/message'
import type { MessageBubbleGestures } from '@/composables/useMessageBubbleMenu'

export interface MessageBubbleContentProps {
  message: MessageItem
  gestures: MessageBubbleGestures
}
