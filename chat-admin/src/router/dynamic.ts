import type { Router, RouteRecordRaw } from 'vue-router'

import { type PermissionItem } from '@/api/permission'

import { routeNameFromCode } from '@/utils/permission'

import { componentMap } from './routes'
import { PermissionType } from '@/constants/permission'

export const HOME_PATH = '/home'

const registeredDynamicNames = new Set<string>()

function buildDynamicRoutes(permissions: PermissionItem[]): RouteRecordRaw[] {
  const routes: RouteRecordRaw[] = []
  for (const p of permissions) {
    if (p.type !== PermissionType.Menu || p.status !== 1 || !p.component || !p.path) continue
    const load = componentMap[p.component]
    if (!load) continue
    routes.push({
      path: p.path,
      name: routeNameFromCode(p.code),
      component: load,
      meta: {
        title: p.name,
        permission: p.code,
        hideInMenu: !!p.hideInMenu,
      },
    } as RouteRecordRaw)
  }
  return routes
}

function routeNeedsRefresh(
  existing: {
    path: string
    meta: Record<string, unknown>
    components?: Record<string, unknown> | null
  },
  desired: RouteRecordRaw,
): boolean {
  const em = existing.meta || {}
  const dm = (desired.meta || {}) as Record<string, unknown>
  const existingComp = existing.components?.default
  return (
    existing.path !== desired.path ||
    existingComp !== desired.component ||
    em.title !== dm.title ||
    em.permission !== dm.permission
  )
}

export function syncDynamicRoutes(router: Router, permissions: PermissionItem[]): boolean {
  const desired = buildDynamicRoutes(permissions)
  const desiredNames = new Set(desired.map((r) => r.name as string))

  registeredDynamicNames.forEach((name) => {
    if (!desiredNames.has(name) && router.hasRoute(name)) {
      router.removeRoute(name)
      registeredDynamicNames.delete(name)
    }
  })

  let added = false
  desired.forEach((r) => {
    const name = r.name as string
    if (router.hasRoute(name)) {
      const existing = router.getRoutes().find((route) => route.name === name)
      if (existing && routeNeedsRefresh(existing, r)) {
        router.removeRoute(name)
        registeredDynamicNames.delete(name)
      } else {
        return
      }
    }
    router.addRoute('Layout', r)
    registeredDynamicNames.add(name)
    added = true
  })
  return added
}

export function resetDynamicRoutes(router: Router) {
  registeredDynamicNames.forEach((name) => {
    if (router.hasRoute(name)) router.removeRoute(name)
  })
  registeredDynamicNames.clear()
}
