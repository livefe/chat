export type LocationPickPurpose = 'chat' | 'publish' | 'room' | 'share'

const LOCATION_PICK_PATH = '/location/pick'

const PURPOSES: LocationPickPurpose[] = ['chat', 'publish', 'room', 'share']

export function isLocationPickPurpose(value: unknown): value is LocationPickPurpose {
  return typeof value === 'string' && PURPOSES.includes(value as LocationPickPurpose)
}

export function locationPickRoute(purpose: LocationPickPurpose) {
  return { path: LOCATION_PICK_PATH, query: { for: purpose } }
}

export function isLocationPickNavigation(to: { path?: string; query?: Record<string, unknown> }) {
  return to.path === LOCATION_PICK_PATH
}
