import { buildClientContextQuery } from '@/utils/clientContext'

export interface ImWSMessage {
  type: string
  content?: string
  clientMsgId?: string
  id?: number
  conversationId?: number
  data?: unknown
}

export type ImWsState = 'connecting' | 'open' | 'closed' | 'failed'

type Handler = (msg: ImWSMessage) => void

const DEFAULT_MAX_RECONNECT = 20
const MAX_BACKOFF_MS = 15000

function looksLikeAuthFailure(ev: CloseEvent): boolean {
  if (ev.code === 4401 || ev.code === 1008) return true
  const reason = (ev.reason || '').toLowerCase()
  return (
    reason.includes('token') ||
    reason.includes('unauthorized') ||
    reason.includes('auth') ||
    reason.includes('expired')
  )
}

export class ImWSClient {
  private socket: WebSocket | null = null
  private timer: number | null = null
  private heartbeat: number | null = null
  private reconnectAttempt = 0
  private manualClose = false
  private authErrorFired = false
  private state: ImWsState = 'closed'

  constructor(
    private opts: {
      url: () => string
      token: () => string
      onMessage?: Handler
      onOpen?: () => void
      onStateChange?: (state: ImWsState) => void
      onAuthError?: () => void

      maxReconnectAttempts?: number
    },
  ) {}

  getState(): ImWsState {
    return this.state
  }

  isFailed(): boolean {
    return this.state === 'failed'
  }

  connect() {
    this.manualClose = false
    this.authErrorFired = false
    this.reconnectAttempt = 0
    this.clearTimers()
    this.setState('connecting')
    this.open()
  }

  reconnect() {
    this.manualClose = false
    this.authErrorFired = false
    this.reconnectAttempt = 0
    this.clearTimers()
    if (this.socket) {
      const prev = this.socket
      this.socket = null
      try {
        prev.close()
      } catch {}
    }
    this.setState('connecting')
    this.open()
  }

  close() {
    this.manualClose = true
    this.clearTimers()
    if (this.socket) {
      try {
        this.socket.close()
      } catch {}
      this.socket = null
    }
    this.setState('closed')
  }

  send(payload: Record<string, unknown>) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(payload))
      return true
    }
    return false
  }

  private setState(next: ImWsState) {
    if (this.state === next) return
    this.state = next
    this.opts.onStateChange?.(next)
  }

  private maxAttempts() {
    return this.opts.maxReconnectAttempts ?? DEFAULT_MAX_RECONNECT
  }

  private fireAuthErrorOnce() {
    if (this.authErrorFired) return
    this.authErrorFired = true
    this.opts.onAuthError?.()
  }

  private open() {
    const token = this.opts.token()
    const base = this.opts.url()
    if (!token || !base) {
      this.scheduleReconnect(1000)
      return
    }
    this.setState('connecting')
    const sep = base.includes('?') ? '&' : '?'
    const url = `${base}${sep}${buildClientContextQuery()}`
    const ws = new WebSocket(url, ['bearer', token])
    this.socket = ws
    ws.onopen = () => {
      if (this.socket !== ws) return
      this.reconnectAttempt = 0
      this.startHeartbeat()
      this.setState('open')
      this.opts.onOpen?.()
    }
    ws.onmessage = (ev) => {
      if (this.socket !== ws) return
      try {
        this.opts.onMessage?.(JSON.parse(ev.data) as ImWSMessage)
      } catch {}
    }
    ws.onclose = (ev) => {
      if (this.socket !== ws) return
      this.socket = null
      this.clearHeartbeat()
      if (this.manualClose) {
        this.setState('closed')
        return
      }
      if (looksLikeAuthFailure(ev)) {
        this.fireAuthErrorOnce()
        this.setState('failed')
        return
      }
      this.setState('closed')
      this.scheduleReconnect()
    }
    ws.onerror = () => {}
  }

  private startHeartbeat() {
    this.clearHeartbeat()
    this.heartbeat = window.setInterval(() => {
      this.send({ type: 'ping' })
    }, 25000)
  }

  private clearHeartbeat() {
    if (this.heartbeat) {
      window.clearInterval(this.heartbeat)
      this.heartbeat = null
    }
  }

  private clearTimers() {
    this.clearHeartbeat()
    if (this.timer) {
      window.clearTimeout(this.timer)
      this.timer = null
    }
  }

  private scheduleReconnect(delay?: number) {
    if (this.manualClose) return
    if (this.reconnectAttempt >= this.maxAttempts()) {
      this.setState('failed')
      return
    }
    this.clearTimers()

    this.setState('connecting')
    const ms = delay ?? Math.min(1000 * 2 ** this.reconnectAttempt, MAX_BACKOFF_MS)
    this.reconnectAttempt += 1
    this.timer = window.setTimeout(() => this.open(), ms)
  }
}

export function buildAdminWsUrl(): string {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/api/admin/ws`
}
