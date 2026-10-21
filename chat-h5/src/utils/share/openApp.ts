import {
  APP_LINK_SCHEME,
  SHARE_FROM,
  SHARE_QUERY,
  SHARE_UTM_SOURCE,
  appDeepLink,
  normalizeAppPath,
  shareResourcePath,
  type ShareChannel,
  type ShareResource,
} from '@/constants/shareLinks'

const ANDROID_PACKAGE = 'com.nearby.chat'

function originBase(): string {
  const origin = window.location.origin
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  return `${origin}${base}`.replace(/([^:]\/)\/+/g, '$1')
}

function buildAppUrl(path: string): string {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${originBase()}${normalizedPath}`.replace(/([^:]\/)\/+/g, '$1')
}

export function buildShareUrl(
  resourceOrPath: ShareResource | string,
  channel: ShareChannel = 'link',
): string {
  const path =
    typeof resourceOrPath === 'string'
      ? normalizeAppPath(resourceOrPath)
      : shareResourcePath(resourceOrPath)
  const url = new URL(buildAppUrl(path.split('?')[0] ?? path))
  if (
    typeof resourceOrPath !== 'string' &&
    resourceOrPath.type === 'download' &&
    resourceOrPath.continuePath
  ) {
    url.searchParams.set(SHARE_QUERY.continue, normalizeAppPath(resourceOrPath.continuePath))
  } else if (typeof resourceOrPath === 'string' && path.includes('?')) {
    const q = new URLSearchParams(path.split('?')[1] || '')
    q.forEach((v, k) => url.searchParams.set(k, v))
  }
  url.searchParams.set(SHARE_QUERY.from, SHARE_FROM.share)
  url.searchParams.set(SHARE_QUERY.utmSource, SHARE_UTM_SOURCE)
  url.searchParams.set(SHARE_QUERY.utmMedium, channel)
  return url.toString()
}

export function buildDownloadUrl(continuePath: string, channel: ShareChannel = 'download'): string {
  return buildShareUrl({ type: 'download', continuePath: normalizeAppPath(continuePath) }, channel)
}

function isAndroid(): boolean {
  return /Android/i.test(navigator.userAgent)
}

function isWeChat(): boolean {
  return /MicroMessenger/i.test(navigator.userAgent)
}

function isIOS(): boolean {
  return /iPhone|iPad|iPod/i.test(navigator.userAgent)
}

function androidIntentUrl(resource: ShareResource, fallbackHttps: string): string {
  const deep = appDeepLink(resource).replace(`${APP_LINK_SCHEME}://`, '')
  const [hostAndPath, query = ''] = deep.split('?')
  const parts = [
    `intent://${hostAndPath}#Intent`,
    `scheme=${APP_LINK_SCHEME}`,
    `package=${ANDROID_PACKAGE}`,
    `S.browser_fallback_url=${encodeURIComponent(fallbackHttps)}`,
  ]
  if (query) {
    for (const [k, v] of new URLSearchParams(query)) {
      parts.push(`S.${k}=${encodeURIComponent(v)}`)
    }
  }
  parts.push('end')
  return parts.join(';')
}

export function tryOpenApp(resource: ShareResource, opts?: { fallbackHttps?: string }): void {
  if (isWeChat()) return

  const path = shareResourcePath(resource)
  const fallback = opts?.fallbackHttps || buildShareUrl(resource, 'open_app')
  const schemeUrl = appDeepLink(resource)

  if (isAndroid()) {
    window.location.href = androidIntentUrl(resource, fallback)
    return
  }

  const start = Date.now()
  window.location.href = schemeUrl
  window.setTimeout(() => {
    if (Date.now() - start < 1600 && !document.hidden && isIOS()) {
      window.location.href = buildDownloadUrl(path.split('?')[0] ?? path, 'download')
    }
  }, 1200)
}
