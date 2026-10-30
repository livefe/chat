import {
  createRouter,
  createWebHistory,
  type LocationQueryRaw,
  type RouteLocationNormalized,
} from 'vue-router'

import { useEmployeeStore } from '@/stores/employee'
import { safeRedirect } from '@/utils/navigation'

import { HOME_PATH, syncDynamicRoutes } from './dynamic'
import { baseRoutes } from './routes'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: baseRoutes,
})

function ensureCatchAll() {
  if (router.hasRoute('NotFound')) return
  router.addRoute('Layout', {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/shell/NotFoundPage.vue'),
    meta: { title: '页面不存在' },
  })
}

async function ensurePermissions(employee: ReturnType<typeof useEmployeeStore>) {
  if (employee.permissionsLoaded) return true
  try {
    await employee.loadPermissions()
    return true
  } catch {
    employee.logout()
    return false
  }
}

function rematchCurrent(to: RouteLocationNormalized) {
  return {
    path: to.path,
    query: to.query as LocationQueryRaw,
    hash: to.hash,
    replace: true as const,
  }
}

export function loginLocation(fullPath?: string) {
  const redirect = safeRedirect(fullPath)
  return {
    path: '/login',
    query: redirect && redirect !== HOME_PATH && redirect !== '/' ? { redirect } : undefined,
  }
}

router.beforeEach(async (to) => {
  const employee = useEmployeeStore()

  if (!employee.isAuthenticated) {
    return to.path === '/login' ? true : loginLocation(to.fullPath)
  }

  if (to.path === '/login') {
    if (to.query.switch === '1') {
      employee.logout()
      return { path: '/login', query: { redirect: safeRedirect(to.query.redirect) }, replace: true }
    }
    if (!(await ensurePermissions(employee))) return true
    ensureCatchAll()
    return { path: safeRedirect(to.query.redirect) ?? HOME_PATH, replace: true }
  }

  if (!(await ensurePermissions(employee))) return loginLocation(to.fullPath)

  syncDynamicRoutes(router, employee.permissions)
  ensureCatchAll()

  if (to.path === '/') {
    return { path: HOME_PATH, replace: true }
  }

  if (to.path === '/403') {
    return true
  }

  if (!to.matched.length || to.name === 'NotFound') {
    const resolved = router.resolve({ path: to.path, query: to.query, hash: to.hash })
    if (resolved.matched.length && resolved.name !== to.name) {
      return rematchCurrent(to)
    }
  }

  const permission = to.meta.permission as string | undefined
  if (permission && !employee.hasPermission(permission)) {
    try {
      await employee.loadPermissions()
      syncDynamicRoutes(router, employee.permissions)
      if (employee.hasPermission(permission)) return true
    } catch {}
    return {
      path: '/403',
      query: {
        from: to.fullPath,
        title: typeof to.meta.title === 'string' ? to.meta.title : undefined,
        permission,
      },
      replace: true,
    }
  }
})

export default router
