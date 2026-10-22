export const AD_STATUS = {
  Draft: 0,
  Pending: 1,
  Active: 2,
  Rejected: 3,
  Paused: 4,
  Ended: 5,
} as const

export const AD_STATUS_OPTIONS = [
  { value: AD_STATUS.Draft, label: '草稿' },
  { value: AD_STATUS.Pending, label: '待审核' },
  { value: AD_STATUS.Active, label: '投放中' },
  { value: AD_STATUS.Rejected, label: '已驳回' },
  { value: AD_STATUS.Paused, label: '已暂停' },
  { value: AD_STATUS.Ended, label: '已结束' },
]

export const AD_LANDING_OPTIONS = [
  { value: 'url', label: '外链' },
  { value: 'post', label: '帖子' },
  { value: 'room', label: '聊天区' },
]

export function formatAdCtr(ctr?: number) {
  if (ctr == null || Number.isNaN(ctr)) return '0%'
  return `${(ctr * 100).toFixed(2)}%`
}
