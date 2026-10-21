export const APP_VERSION = '1.0.0'

export const CLIENT_PLATFORM = 'h5' as const

export const ClientHeaders = {
  REQUEST_ID: 'X-Request-ID',
  CLIENT_PLATFORM: 'X-Client-Platform',
  CLIENT_VERSION: 'X-Client-Version',
  DEVICE_ID: 'X-Device-ID',
} as const

const DEVICE_ID_KEY = 'nearby_h5_device_id'

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

export function getOrCreateDeviceId(): string {
  const existing = localStorage.getItem(DEVICE_ID_KEY)
  if (existing) return existing
  const id = newUUID()
  localStorage.setItem(DEVICE_ID_KEY, id)
  return id
}

export function newRequestId(): string {
  return newUUID()
}
