export const APP_LINK_SCHEME = String(import.meta.env.VITE_APP_SCHEME || 'nearby').replace(
  /:\/\/*$/,
  '',
)

export const SHARE_QUERY = {
  from: 'from',
  continue: 'continue',
  utmSource: 'utm_source',
  utmMedium: 'utm_medium',
} as const

export const SHARE_FROM = {
  share: 'share',
  app: 'app',
} as const

export const SHARE_UTM_SOURCE = 'share'

export type ShareChannel =
  'link' | 'copy' | 'system' | 'wechat' | 'xhs' | 'douyin' | 'qr' | 'open_app' | 'download'

export type ShareResource =
  | { type: 'post'; id: number }
  | { type: 'room'; id: number }
  | { type: 'user'; id: number }
  | { type: 'messages'; id: number }
  | { type: 'download'; continuePath?: string }

export function shareResourcePath(resource: ShareResource): string {
  switch (resource.type) {
    case 'post':
      return `/post/${resource.id}`
    case 'room':
      return `/room/${resource.id}`
    case 'user':
      return `/user/${resource.id}`
    case 'messages':
      return `/messages/${resource.id}`
    case 'download': {
      if (!resource.continuePath) return '/download'
      const q = new URLSearchParams()
      q.set(SHARE_QUERY.continue, normalizeAppPath(resource.continuePath))
      return `/download?${q.toString()}`
    }
  }
}

export function normalizeAppPath(raw: string): string {
  const path = raw.trim()
  if (!path.startsWith('/')) return '/'
  if (path.startsWith('//') || path.includes('://')) return '/'
  return path.split('#')[0] || '/'
}

export function parseShareResourcePath(pathWithQuery: string): ShareResource | null {
  const raw = normalizeAppPath(pathWithQuery)
  const parts = raw.split('?')
  const pathname = parts[0] ?? '/'
  const query = parts[1] ?? ''
  const post = pathname.match(/^\/post\/(\d+)$/)
  if (post) return { type: 'post', id: Number(post[1]) }
  const room = pathname.match(/^\/room\/(\d+)$/)
  if (room) return { type: 'room', id: Number(room[1]) }
  const user = pathname.match(/^\/user\/(\d+)$/)
  if (user) return { type: 'user', id: Number(user[1]) }
  const messages = pathname.match(/^\/messages\/(\d+)$/)
  if (messages) return { type: 'messages', id: Number(messages[1]) }
  if (pathname === '/download' || pathname.startsWith('/download')) {
    const continuePath = new URLSearchParams(query).get(SHARE_QUERY.continue) || undefined
    return {
      type: 'download',
      continuePath: continuePath ? normalizeAppPath(continuePath) : undefined,
    }
  }
  return null
}

export function appDeepLink(resource: ShareResource): string {
  switch (resource.type) {
    case 'post':
      return `${APP_LINK_SCHEME}://post/${resource.id}`
    case 'room':
      return `${APP_LINK_SCHEME}://room/${resource.id}`
    case 'user':
      return `${APP_LINK_SCHEME}://user/${resource.id}`
    case 'messages':
      return `${APP_LINK_SCHEME}://messages/${resource.id}`
    case 'download': {
      const base = `${APP_LINK_SCHEME}://download`
      if (!resource.continuePath) return base
      return `${base}?${SHARE_QUERY.continue}=${encodeURIComponent(normalizeAppPath(resource.continuePath))}`
    }
  }
}

export function isShareTraffic(query: Record<string, unknown> | undefined | null): boolean {
  return String(query?.[SHARE_QUERY.from] || '') === SHARE_FROM.share
}
