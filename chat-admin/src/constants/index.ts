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

export const PAGE_SIZES = [10, 20, 50, 100]

export const PAGINATION_LAYOUT = 'total, sizes, prev, pager, next, jumper'
