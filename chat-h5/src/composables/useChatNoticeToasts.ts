import { onScopeDispose } from 'vue'
import { showToast } from 'vant'

import { useChatStore, type ChatNotice } from '@/stores/chat'
import { showCatchError } from '@/utils/toast'

function showNotice(notice: ChatNotice) {
  switch (notice.type) {
    case 'sendTimeout':
      showToast('消息发送超时，请检查网络')
      return
    case 'serverError':
      showToast(notice.message)
      return
    case 'syncFailed':
      showCatchError(notice.error, '同步消息失败')
      return
    case 'unreadFailed':
      showCatchError(notice.error, '获取未读失败')
  }
}

export function useChatNoticeToasts() {
  const off = useChatStore().onNotice(showNotice)
  onScopeDispose(off)
}
