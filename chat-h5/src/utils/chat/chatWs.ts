import type { WsScope } from '@/constants/chat'

export interface WSMessage {
  type: string
  scope?: WsScope
  roomId?: number
  postId?: number
  onlineCount?: number
  toUserId?: number
  lat?: number
  lng?: number
  contentType?: number
  content?: string
  replyToId?: number
  clientMsgId?: string
  id?: number
  data?: unknown
}

type Handler = (msg: WSMessage) => void

export interface WSOptions {
  url: () => string

  protocols?: () => string[] | undefined
  onOpen?: () => void
  onClose?: (ev: CloseEvent) => boolean | void
  onMessage?: Handler
}

export class WSClient {
  private socket: WebSocket | null = null
  private timer: number | null = null
  private heartbeat: number | null = null
  private reconnectAttempt = 0
  private manualClose = false

  constructor(private opts: WSOptions) {}

  connect() {
    this.manualClose = false
    this.open()
  }

  retryNow() {
    if (this.manualClose) return
    if (this.socket) {
      if (this.socket.readyState === WebSocket.OPEN) return
      if (this.socket.readyState === WebSocket.CONNECTING) return
      if (this.socket.readyState === WebSocket.CLOSING) {
        this.safeCloseSocket(this.socket)
        return
      }
      this.socket = null
    }
    if (this.timer) {
      window.clearTimeout(this.timer)
      this.timer = null
    }
    this.reconnectAttempt = 0
    this.open()
  }

  private open() {
    try {
      if (this.socket) {
        if (
          this.socket.readyState === WebSocket.OPEN ||
          this.socket.readyState === WebSocket.CONNECTING
        ) {
          return
        }
        this.safeCloseSocket(this.socket)
        this.socket = null
      }

      const url = this.opts.url()
      if (!url) {
        this.scheduleReconnect(1000)
        return
      }
      const protocols = this.opts.protocols?.()
      const ws = protocols?.length ? new WebSocket(url, protocols) : new WebSocket(url)
      this.socket = ws
      ws.onopen = () => {
        if (this.socket !== ws) return
        this.reconnectAttempt = 0
        this.startHeartbeat()
        this.opts.onOpen?.()
      }
      ws.onmessage = (ev) => {
        if (this.socket !== ws) return
        try {
          const msg = JSON.parse(ev.data) as WSMessage
          this.opts.onMessage?.(msg)
        } catch {}
      }
      ws.onclose = (ev) => {
        if (this.socket !== ws) return
        this.socket = null
        this.stopHeartbeat()
        let shouldReconnect = true
        const ret = this.opts.onClose?.(ev)
        if (ret === false) shouldReconnect = false
        if (!this.manualClose && shouldReconnect) this.scheduleReconnect()
      }
      ws.onerror = () => {
        if (this.socket !== ws) return
        ws.close()
      }
    } catch {
      this.scheduleReconnect()
    }
  }

  private scheduleReconnect(delayOverride?: number) {
    const delay = delayOverride ?? Math.min(1000 * 2 ** this.reconnectAttempt, 15000)
    this.reconnectAttempt += 1
    if (this.timer) window.clearTimeout(this.timer)
    this.timer = window.setTimeout(() => this.open(), delay)
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeat = window.setInterval(() => {
      this.send({ type: 'ping' })
    }, 25000)
  }

  private stopHeartbeat() {
    if (this.heartbeat) {
      window.clearInterval(this.heartbeat)
      this.heartbeat = null
    }
  }

  send(msg: WSMessage) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(msg))
      return true
    }
    return false
  }

  isOpen() {
    return this.socket?.readyState === WebSocket.OPEN
  }

  close() {
    this.manualClose = true
    this.stopHeartbeat()
    if (this.timer) {
      window.clearTimeout(this.timer)
      this.timer = null
    }
    if (this.socket) {
      this.safeCloseSocket(this.socket)
    }
    this.socket = null
  }

  private safeCloseSocket(ws: WebSocket) {
    try {
      ws.close()
    } catch {}
  }
}
