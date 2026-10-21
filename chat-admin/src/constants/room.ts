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
