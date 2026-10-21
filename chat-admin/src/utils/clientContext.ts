const APP_VERSION = '1.0.0'
const CLIENT_PLATFORM = 'admin' as const

const ClientHeaders = {
  REQUEST_ID: 'X-Request-ID',
  CLIENT_PLATFORM: 'X-Client-Platform',
  CLIENT_VERSION: 'X-Client-Version',
  DEVICE_ID: 'X-Device-ID',
} as const

const DEVICE_ID_KEY = 'chat_admin_device_id'

function newUUID(): string {
  const c = globalThis.crypto
  if (c && typeof c.randomUUID === 'function') return c.randomUUID()
  if (c && typeof c.getRandomValues === 'function') {
    const bytes = new Uint8Array(16)
    c.getRandomValues(bytes)
    const b6 = bytes[6] ?? 0
    const b8 = bytes[8] ?? 0
    bytes[6] = (b6 & 0x0f) | 0x40
    bytes[8] = (b8 & 0x3f) | 0x80
    const hex = [...bytes].map((b) => b.toString(16).padStart(2, '0')).join('')
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
  }
  return `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`.replace(/[xy]/g, (ch) => {
    const n = (Math.random() * 16) | 0
    const v = ch === 'x' ? n : (n & 0x3) | 0x8
    return v.toString(16)
  })
}

function getOrCreateDeviceId(): string {
  const existing = localStorage.getItem(DEVICE_ID_KEY)
  if (existing) return existing
  const id = newUUID()
  localStorage.setItem(DEVICE_ID_KEY, id)
  return id
}

function newRequestId(): string {
  return newUUID()
}

export function buildClientContextHeaders(): Record<string, string> {
  return {
    [ClientHeaders.REQUEST_ID]: newRequestId(),
    [ClientHeaders.CLIENT_PLATFORM]: CLIENT_PLATFORM,
    [ClientHeaders.CLIENT_VERSION]: APP_VERSION,
    [ClientHeaders.DEVICE_ID]: getOrCreateDeviceId(),
  }
}

export function buildClientContextQuery(): string {
  const params = new URLSearchParams({
    requestId: newRequestId(),
    clientPlatform: CLIENT_PLATFORM,
    clientVersion: APP_VERSION,
    deviceId: getOrCreateDeviceId(),
  })
  return params.toString()
}
