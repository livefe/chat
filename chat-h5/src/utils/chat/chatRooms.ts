import type { NearbyRoomItem } from '@/api/room'

export function sortRoomsByDistance(rooms: readonly NearbyRoomItem[]): NearbyRoomItem[] {
  return [...rooms].sort((a, b) => a.distance - b.distance)
}
