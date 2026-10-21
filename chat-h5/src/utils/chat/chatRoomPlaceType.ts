const RoomPlaceType = {
  Community: 1,
  School: 2,
  Business: 3,
  Scenic: 4,
  Campus: 5,
  Transit: 6,
  Other: 7,
} as const

export type RoomPlaceTypeValue = (typeof RoomPlaceType)[keyof typeof RoomPlaceType]

export const ROOM_PLACE_TYPE_OPTIONS: {
  value: RoomPlaceTypeValue
  label: string
  mark: string
}[] = [
  { value: RoomPlaceType.Community, label: '小区', mark: '小' },
  { value: RoomPlaceType.School, label: '学校', mark: '校' },
  { value: RoomPlaceType.Business, label: '商圈', mark: '商' },
  { value: RoomPlaceType.Scenic, label: '景点', mark: '景' },
  { value: RoomPlaceType.Campus, label: '园区', mark: '园' },
  { value: RoomPlaceType.Transit, label: '交通枢纽', mark: '交' },
  { value: RoomPlaceType.Other, label: '其他', mark: '其' },
]

export function roomPlaceTypeLabel(placeType: number | undefined | null): string {
  const hit = ROOM_PLACE_TYPE_OPTIONS.find((o) => o.value === placeType)
  return hit?.label || '其他'
}

export function roomPlaceTypeMark(placeType: number | undefined | null): string {
  const hit = ROOM_PLACE_TYPE_OPTIONS.find((o) => o.value === placeType)
  return hit?.mark || '其'
}
