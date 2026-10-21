export enum PostStatus {
  Pending = 0,
  Published = 1,
  Rejected = 2,
  Disabled = 3,
}

export const POST_STATUS_OPTIONS = [
  { label: '待审核', value: PostStatus.Pending },
  { label: '发布中', value: PostStatus.Published },
  { label: '驳回', value: PostStatus.Rejected },
  { label: '下架', value: PostStatus.Disabled },
] as const

export const PostCategory = {
  Event: 'event',
  Notice: 'notice',
  Idle: 'idle',
  Help: 'help',
} as const

type PostCategoryValue = (typeof PostCategory)[keyof typeof PostCategory]

export const POST_CATEGORY_OPTIONS: { value: PostCategoryValue; label: string }[] = [
  { value: PostCategory.Event, label: '活动' },
  { value: PostCategory.Notice, label: '公告' },
  { value: PostCategory.Idle, label: '闲置' },
  { value: PostCategory.Help, label: '求助' },
]

export enum SignupStatus {
  Confirmed = 1,
  Waitlisted = 2,
  Cancelled = 3,
}

export enum SignupCheckIn {
  NotChecked = 1,
  Checked = 2,
}
