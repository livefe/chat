export function resolveMediaUrl(url?: string | null): string {
  const u = (url || '').trim()
  if (!u) return ''
  if (/^(https?:|blob:|data:)/i.test(u)) return u

  const path = u.startsWith('/') ? u : `/${u}`
  return `${window.location.origin}${path}`
}

export function collectMediaIds(list: { mediaId?: number }[]): number[] {
  return list.map((x) => x.mediaId).filter((id): id is number => typeof id === 'number' && id > 0)
}

export function resolveMediaUrls(urls?: string[] | null): string[] {
  if (!urls?.length) return []
  return urls.map(resolveMediaUrl).filter(Boolean)
}
