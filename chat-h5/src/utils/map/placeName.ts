export const PLACE_NAME_NEARBY = '附近'

export function realPlaceNameOrEmpty(name: string | undefined | null): string {
  const n = (name || '').trim()
  return !n || n === PLACE_NAME_NEARBY ? '' : n
}

export function displayPlaceName(
  name: string | undefined | null,
  opts: { isCurrentGps: boolean },
): string {
  const real = realPlaceNameOrEmpty(name)
  if (real) return real
  return opts.isCurrentGps ? PLACE_NAME_NEARBY : ''
}
