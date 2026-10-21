import { defineStore } from 'pinia'
import { ref } from 'vue'

import { getUserDetail, refreshTokens, type UserInfo } from '@/api/user'
import { H5StorageKey, readH5AuthStorage } from '@/utils/storage'

const { accessToken: storedAccess, refreshToken: storedRefresh } = readH5AuthStorage()

export const useUserStore = defineStore('user', () => {
  const accessToken = ref<string>(storedAccess)
  const refreshToken = ref<string>(storedRefresh)
  const info = ref<UserInfo | null>(null)

  function setTokens(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    if (access) {
      localStorage.setItem(H5StorageKey.AccessToken, access)
    } else {
      localStorage.removeItem(H5StorageKey.AccessToken)
    }
    if (refresh) {
      localStorage.setItem(H5StorageKey.RefreshToken, refresh)
    } else {
      localStorage.removeItem(H5StorageKey.RefreshToken)
    }
  }

  function setUser(u: UserInfo | null) {
    info.value = u
  }

  let refreshInFlight: Promise<boolean> | null = null

  async function tryRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    if (refreshInFlight) return refreshInFlight
    refreshInFlight = (async () => {
      try {
        const res = await refreshTokens(refreshToken.value)

        setTokens(res.data.accessToken, res.data.refreshToken)
        return true
      } catch {
        setTokens('', '')
        info.value = null
        return false
      } finally {
        refreshInFlight = null
      }
    })()
    return refreshInFlight
  }

  async function fetchDetail() {
    if (!accessToken.value) return null
    try {
      const res = await getUserDetail()
      info.value = res.data
      return res.data
    } catch {
      return null
    }
  }

  async function ensureSession(): Promise<boolean> {
    if (!accessToken.value && refreshToken.value) {
      const ok = await tryRefresh()
      if (!ok) return false
    }
    if (!accessToken.value) return false
    if (info.value) return true

    let detail = await fetchDetail()
    if (detail) return true

    const refreshed = await tryRefresh()
    if (!refreshed) return false
    detail = await fetchDetail()
    return !!detail
  }

  let logoutHandler: () => void = () => {}

  function setLogoutHandler(handler: () => void) {
    logoutHandler = handler
  }

  function logout() {
    setTokens('', '')
    info.value = null
    logoutHandler()
  }

  return {
    accessToken,
    refreshToken,
    info,
    setTokens,
    setUser,
    tryRefresh,
    fetchDetail,
    ensureSession,
    logout,
    setLogoutHandler,
  }
})
