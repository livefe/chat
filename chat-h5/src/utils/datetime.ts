export function pad(n: number): string {
  return String(n).padStart(2, '0')
}

export function toUnixMs(ts: number): number {
  if (!ts) return 0
  return ts * 1000
}

export function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  )
}

function parseStorageDateTime(v: string): Date | null {
  if (!v) return null
  const d = new Date(v.replace(' ', 'T'))
  return Number.isNaN(d.getTime()) ? null : d
}

export function formatDateTimeChinese(v: string, withSeconds = false): string {
  const d = parseStorageDateTime(v)
  if (!d) return ''
  const base = `${d.getFullYear()}年${pad(d.getMonth() + 1)}月${pad(d.getDate())}日 ${pad(d.getHours())}:${pad(d.getMinutes())}`
  return withSeconds ? `${base}:${pad(d.getSeconds())}` : base
}

export function formatDateChinese(v: string): string {
  const d = parseStorageDateTime(v)
  if (!d) return ''
  return `${d.getFullYear()}年${pad(d.getMonth() + 1)}月${pad(d.getDate())}日`
}

export function parseDateTimeToUnix(v: string): number {
  const d = parseStorageDateTime(v)
  return d ? Math.floor(d.getTime() / 1000) : 0
}

export function unixToStorageDateTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(toUnixMs(ts))
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function toStorageDateTime(y: string, m: string, d: string, hh: string, mm: string): string {
  return `${y}-${m.padStart(2, '0')}-${d.padStart(2, '0')} ${hh.padStart(2, '0')}:${mm.padStart(2, '0')}`
}

export function splitDateTime(v: string): { date: string[]; time: string[] } {
  const d = parseStorageDateTime(v) ?? new Date()
  return {
    date: [String(d.getFullYear()), pad(d.getMonth() + 1), pad(d.getDate())],
    time: [pad(d.getHours()), pad(d.getMinutes())],
  }
}
