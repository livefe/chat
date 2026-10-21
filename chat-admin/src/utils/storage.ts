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
