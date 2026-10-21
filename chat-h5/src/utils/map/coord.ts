export function hasCoordinates(lat?: number, lng?: number): boolean {
  if (lat === undefined || lng === undefined) return false
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) return false
  if (lat === 0 && lng === 0) return false
  if (lat < -90 || lat > 90 || lng < -180 || lng > 180) return false
  return true
}
