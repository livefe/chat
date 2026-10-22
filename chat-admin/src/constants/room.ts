export enum RoomStatus {
  Pending = 0,
  Enabled = 1,
  Rejected = 2,
  Disabled = 3,
}

export const ROOM_STATUS_OPTIONS = [
  { label: '待审核', value: RoomStatus.Pending },
  { label: '正常', value: RoomStatus.Enabled },
  { label: '驳回', value: RoomStatus.Rejected },
  { label: '禁用', value: RoomStatus.Disabled },
] as const

export enum RoomSource {
  Admin = 1,
  User = 2,
}

export const ROOM_SOURCE_OPTIONS = [
  { label: '后台', value: RoomSource.Admin },
  { label: '用户', value: RoomSource.User },
] as const

export function roomStatusTag(status: number): 'success' | 'warning' | 'info' | 'danger' {
  if (status === RoomStatus.Enabled) return 'success'
  if (status === RoomStatus.Pending) return 'warning'
  if (status === RoomStatus.Rejected) return 'info'
  return 'danger'
}

const RoomPlaceType = {
  Community: 1,
  School: 2,
  Business: 3,
  Scenic: 4,
  Campus: 5,
  Transit: 6,
  Other: 7,
} as const

type RoomPlaceTypeValue = (typeof RoomPlaceType)[keyof typeof RoomPlaceType]

export const ROOM_PLACE_TYPE_OPTIONS: {
  value: RoomPlaceTypeValue
  label: string
}[] = [
  { value: RoomPlaceType.Community, label: '小区' },
  { value: RoomPlaceType.School, label: '学校' },
  { value: RoomPlaceType.Business, label: '商圈' },
  { value: RoomPlaceType.Scenic, label: '景点' },
  { value: RoomPlaceType.Campus, label: '园区' },
  { value: RoomPlaceType.Transit, label: '交通枢纽' },
  { value: RoomPlaceType.Other, label: '其他' },
]
