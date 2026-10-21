import { ElMessage } from 'element-plus'
import type { Router } from 'vue-router'

import type { PermissionItem } from '@/api/permission'
import { PermissionType, type PermissionCode } from '@/constants/permission'

export function findMenuPath(
  permissions: PermissionItem[],
  code: PermissionCode,
): string | undefined {
  return permissions.find(
    (p) => p.code === code && p.type === PermissionType.Menu && p.status === 1 && p.path,
  )?.path
}

export function pushMenuRoute(
  router: Router,
  permissions: PermissionItem[],
  code: PermissionCode,
  query?: Record<string, string>,
) {
  const path = findMenuPath(permissions, code)
  if (!path) {
    ElMessage.warning('暂无访问权限')
    return
  }
  router.push({ path, query })
}
