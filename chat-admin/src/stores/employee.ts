import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, refreshTokenApi } from '@/api/auth'
import type { EmployeeInfo } from '@/api/employee'
import { getEmployeePermissions, type PermissionItem } from '@/api/permission'
import { resetDynamicRoutes, syncDynamicRoutes } from '@/router/dynamic'
import { clearAdminStorage, readAdminStorage, AdminStorageKey } from '@/utils/storage'

async function getRouter() {
  const { default: router } = await import('@/router')
  return router
}

const stored = readAdminStorage()

export const useEmployeeStore = defineStore('employee', () => {
  const accessToken = ref<string>(stored.accessToken)
  const refreshToken = ref<string>(stored.refreshToken)
  const employeeInfo = ref<EmployeeInfo | null>(stored.employeeInfo as EmployeeInfo | null)
  const permissions = ref<PermissionItem[]>([])

  const permissionsLoaded = ref<boolean>(false)
  let refreshPromise: Promise<boolean> | null = null

  const isAuthenticated = computed(() => !!accessToken.value)
  const permissionCodeSet = computed(
    () => new Set(permissions.value.filter((p) => p.status === 1).map((p) => p.code)),
  )

  function setTokens(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    if (access) {
      localStorage.setItem(AdminStorageKey.AccessToken, access)
    } else {
      localStorage.removeItem(AdminStorageKey.AccessToken)
    }
    if (refresh) {
      localStorage.setItem(AdminStorageKey.RefreshToken, refresh)
    } else {
      localStorage.removeItem(AdminStorageKey.RefreshToken)
    }
  }

  function setEmployeeInfo(info: EmployeeInfo) {
    employeeInfo.value = info
    localStorage.setItem(AdminStorageKey.EmployeeInfo, JSON.stringify(info))
  }

  function setPermissions(list: PermissionItem[]) {
    permissions.value = list
    permissionsLoaded.value = true
  }

  async function tryRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    if (refreshPromise) return refreshPromise
    refreshPromise = (async () => {
      try {
        const { data } = await refreshTokenApi(refreshToken.value)
        setTokens(data.accessToken, data.refreshToken)
        return true
      } catch {
        return false
      } finally {
        refreshPromise = null
      }
    })()
    return refreshPromise
  }

  async function login(mobile: string, code: string) {
    const { data } = await loginApi(mobile, code)
    setTokens(data.accessToken, data.refreshToken)
    setEmployeeInfo(data.employee)
    try {
      await loadPermissions()
    } catch (error) {
      logout()
      throw error
    }
  }

  async function loadPermissions() {
    const { data } = await getEmployeePermissions()
    setPermissions(Array.isArray(data) ? data : [])
    const router = await getRouter()
    syncDynamicRoutes(router, permissions.value)
  }

  function logout() {
    accessToken.value = ''
    refreshToken.value = ''
    employeeInfo.value = null
    permissions.value = []
    permissionsLoaded.value = false
    clearAdminStorage()
    void getRouter().then((router) => resetDynamicRoutes(router))
  }

  function hasPermission(code: string) {
    return permissionCodeSet.value.has(code)
  }

  return {
    accessToken,
    employeeInfo,
    permissions,
    permissionsLoaded,
    isAuthenticated,
    setEmployeeInfo,
    login,
    logout,
    tryRefresh,
    loadPermissions,
    hasPermission,
  }
})
