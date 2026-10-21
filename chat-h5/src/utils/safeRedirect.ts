export function safeRedirectPath(raw: unknown, fallback = '/'): string {
  if (typeof raw !== 'string') return fallback
  let path = raw.trim()
  if (!path) return fallback

  try {
    path = decodeURIComponent(path)
  } catch {
    return fallback
  }
  path = path.trim()

  if (!path.startsWith('/')) return fallback

  if (path.includes('\\')) return fallback
  if (path.startsWith('//')) return fallback
  if (path.includes('://')) return fallback

  if (/^\/[^/]*:/.test(path)) return fallback

  return path
}
