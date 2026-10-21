import { onUnmounted, ref } from 'vue'

import { useEmployeeStore } from '@/stores/employee'
import { ImWSClient, buildAdminWsUrl, type ImWSMessage, type ImWsState } from '@/utils/imWs'

type Options = {
  onMessage: (msg: ImWSMessage) => void
  onOpen?: () => void
}

export function useImConnection(opts: Options) {
  const employeeStore = useEmployeeStore()
  const wsState = ref<ImWsState>('closed')
  let ws: ImWSClient | null = null

  function connect() {
    ws?.close()
    ws = new ImWSClient({
      url: buildAdminWsUrl,
      token: () => employeeStore.accessToken,
      onMessage: opts.onMessage,
      onOpen: opts.onOpen,
      onStateChange: (s) => {
        wsState.value = s
      },
      onAuthError: async () => {
        const ok = await employeeStore.tryRefresh()
        if (ok) ws?.reconnect()
        else employeeStore.logout()
      },
    })
    ws.connect()
  }

  function reconnect() {
    if (!ws) {
      connect()
      return
    }
    ws.reconnect()
  }

  function close() {
    ws?.close()
    ws = null
    wsState.value = 'closed'
  }

  function isOpen() {
    return wsState.value === 'open'
  }

  function send(payload: Record<string, unknown>) {
    return ws?.send(payload) ?? false
  }

  onUnmounted(close)

  return {
    wsState,
    connect,
    reconnect,
    close,
    isOpen,
    send,
  }
}
