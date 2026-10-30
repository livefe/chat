export const MOBILE_PATTERN = /^1[3-9]\d{9}$/

export const MOBILE_RULE = {
  pattern: MOBILE_PATTERN,
  message: '请输入正确的手机号',
  trigger: 'blur' as const,
}

export const STATUS_OPTIONS = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
] as const

export const DEFAULT_PAGE_SIZE = 20

export const FILTER_ALL = -1

export const PAGE_SIZES = [10, 20, 50, 100]

export const PAGINATION_LAYOUT = 'total, sizes, prev, pager, next, jumper'

export const DATETIME_FORMAT = 'YYYY-MM-DD HH:mm:ss'

export const DATE_RANGE_DEFAULT_TIME: [Date, Date] = [
  new Date(2000, 0, 1, 0, 0, 0),
  new Date(2000, 0, 1, 23, 59, 59),
]

function startOfDay(d: Date) {
  const out = new Date(d)
  out.setHours(0, 0, 0, 0)
  return out
}

function endOfDay(d: Date) {
  const out = new Date(d)
  out.setHours(23, 59, 59, 0)
  return out
}

function daysAgo(n: number) {
  const d = new Date()
  d.setDate(d.getDate() - n)
  return d
}

export const DATE_RANGE_SHORTCUTS = [
  { text: '今天', value: () => [startOfDay(new Date()), endOfDay(new Date())] },
  { text: '昨天', value: () => [startOfDay(daysAgo(1)), endOfDay(daysAgo(1))] },
  { text: '近 7 天', value: () => [startOfDay(daysAgo(6)), endOfDay(new Date())] },
  { text: '近 30 天', value: () => [startOfDay(daysAgo(29)), endOfDay(new Date())] },
  {
    text: '本月',
    value: () => {
      const now = new Date()
      return [startOfDay(new Date(now.getFullYear(), now.getMonth(), 1)), endOfDay(now)]
    },
  },
]
