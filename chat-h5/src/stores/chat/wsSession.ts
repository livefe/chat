import type { Ref } from 'vue'

import { notifyAuthLost } from '@/api/request'
import {
  APP_VERSION,
  CLIENT_PLATFORM,
  getOrCreateDeviceId,
  newRequestId,
} from '@/constants/appVersion'
import { WSClient, type WSMessage } from '@/utils/chat/chatWs'

export interface WsSessionDeps {
  connected: Ref<boolean>
  getToken: () => string
  tryRefresh: () => Promise<boolean>
  onOpen: () => void
  onMessage: (env: WSMessage) => void
  onBeforeReconnect: () => void
  onDisconnectCleanup: (opts?: { clearSession?: boolean }) => void
  reportLocation: () => void
}

export function createWsSession(deps: WsSessionDeps) {
  let client: WSClient | null = null

  function getClient() {
    return client
  }

  function connect() {
    if (client) return
    client = new WSClient({
      url: () => {
        if (!deps.getToken()) return ''
        const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
        const q = new URLSearchParams({
          requestId: newRequestId(),
          clientPlatform: CLIENT_PLATFORM,
          clientVersion: APP_VERSION,
          deviceId: getOrCreateDeviceId(),
        })
        return `${proto}://${window.location.host}/api/ws?${q.toString()}`
      },
      protocols: () => {
        const token = deps.getToken()
        if (!token) return undefined

        return ['bearer', token]
      },
      onOpen: () => {
        deps.connected.value = true
        deps.onOpen()
      },
      onClose: (ev) => {
        deps.connected.value = false

        if (ev.code === 4401) {
          void deps.tryRefresh().then((ok) => {
            if (ok) {
              client?.retryNow()
            } else {
              client?.close()
              client = null
              notifyAuthLost()
            }
          })
          return false
        }
      },
      onMessage: deps.onMessage,
    })
    client.connect()
  }

  function ensureAlive() {
    if (!deps.getToken()) return
    if (!client) {
      connect()
      return
    }
    if (!client.isOpen()) {
      client.retryNow()
      return
    }

    deps.reportLocation()
  }

  function waitForOpen(timeoutMs = 8000): Promise<boolean> {
    if (client?.isOpen()) return Promise.resolve(true)
    ensureAlive()
    return new Promise((resolve) => {
      const start = Date.now()
      const tick = () => {
        if (client?.isOpen()) {
          resolve(true)
          return
        }
        if (Date.now() - start >= timeoutMs) {
          resolve(false)
          return
        }
        window.setTimeout(tick, 150)
      }
      window.setTimeout(tick, 150)
    })
  }

  function disconnect(opts?: { clearSession?: boolean }) {
    deps.onDisconnectCleanup(opts)
    client?.close()
    client = null
    deps.connected.value = false
  }

  return {
    getClient,
    connect,
    ensureAlive,
    waitForOpen,
    disconnect,
  }
}
