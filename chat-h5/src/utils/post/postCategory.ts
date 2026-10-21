export const PostCategory = {
  Event: 'event',
  Notice: 'notice',
  Idle: 'idle',
  Help: 'help',
} as const

export type PostCategoryValue = (typeof PostCategory)[keyof typeof PostCategory]

export const POST_CATEGORY_OPTIONS: {
  value: PostCategoryValue
  label: string
  hint: string
}[] = [
  { value: PostCategory.Event, label: '活动', hint: '组局、报名' },
  { value: PostCategory.Notice, label: '公告', hint: '通知、提醒' },
  { value: PostCategory.Idle, label: '闲置', hint: '转让、赠送' },
  { value: PostCategory.Help, label: '求助', hint: '帮忙、悬赏' },
]

export function postCategoryOption(category: string) {
  return POST_CATEGORY_OPTIONS.find((x) => x.value === category)
}
