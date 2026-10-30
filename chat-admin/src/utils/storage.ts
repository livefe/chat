export const AdminStorageKey = {
  AccessToken: 'admin_access_token',
  RefreshToken: 'admin_refresh_token',
  EmployeeInfo: 'employeeInfo',
} as const

const ADMIN_STORAGE_KEYS = Object.values(AdminStorageKey)

export function clearAdminStorage() {
  ADMIN_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key))
}

function readJsonStorage<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    if (raw === null) return fallback
    return JSON.parse(raw) as T
  } catch {
    localStorage.removeItem(key)
    return fallback
  }
}

export function readAdminStorage() {
  try {
    return {
      accessToken: localStorage.getItem(AdminStorageKey.AccessToken) || '',
      refreshToken: localStorage.getItem(AdminStorageKey.RefreshToken) || '',
      employeeInfo: readJsonStorage<EmployeeInfoLike | null>(AdminStorageKey.EmployeeInfo, null),
    }
  } catch {
    clearAdminStorage()
    return {
      accessToken: '',
      refreshToken: '',
      employeeInfo: null,
    }
  }
}

type EmployeeInfoLike = Record<string, unknown>

export interface RecentVisit {
  path: string
  title: string
}

const RECENT_VISITS_PREFIX = 'admin_recent_visits:'
const RECENT_VISITS_MAX = 6

export function readRecentVisits(employeeId: number): RecentVisit[] {
  const list = readJsonStorage<unknown>(RECENT_VISITS_PREFIX + employeeId, [])
  if (!Array.isArray(list)) return []
  return list.filter(
    (v): v is RecentVisit =>
      !!v && typeof v.path === 'string' && typeof v.title === 'string' && v.path.startsWith('/'),
  )
}

export function pushRecentVisit(employeeId: number, visit: RecentVisit) {
  const list = readRecentVisits(employeeId).filter((v) => v.path !== visit.path)
  list.unshift(visit)
  localStorage.setItem(
    RECENT_VISITS_PREFIX + employeeId,
    JSON.stringify(list.slice(0, RECENT_VISITS_MAX)),
  )
}
