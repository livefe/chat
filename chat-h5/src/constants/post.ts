export const PostFeeType = {
  None: 0,
  Charge: 1,
  Reward: 2,
} as const

export const PostStatus = {
  Pending: 0,
  Published: 1,
  Rejected: 2,
  Disabled: 3,
} as const

export const SignupStatus = {
  Confirmed: 1,
  Waitlisted: 2,
  Cancelled: 3,
} as const

export const SignupCheckIn = {
  NotChecked: 1,
  Checked: 2,
} as const
