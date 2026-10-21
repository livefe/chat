import { PermissionType } from '@/constants/permission'

export const permissionTypeOptions = [
  { label: '菜单', value: PermissionType.Menu },
  { label: '按钮', value: PermissionType.Button },
  { label: '接口', value: PermissionType.API },
  { label: '数据', value: PermissionType.Data },
]

export const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE']

export function permissionTypeLabel(type: number): string {
  return permissionTypeOptions.find((o) => o.value === type)?.label ?? '未知'
}

export function permissionTypeTag(type: number): 'primary' | 'warning' | 'info' | 'success' {
  switch (type) {
    case PermissionType.Menu:
      return 'primary'
    case PermissionType.Button:
      return 'warning'
    case PermissionType.API:
      return 'info'
    default:
      return 'success'
  }
}

export function routeNameFromCode(code: string): string {
  return code.replace(/:/g, '-')
}

export function compareBySortThenId(
  a: { sort: number; id: number },
  b: { sort: number; id: number },
): number {
  return a.sort !== b.sort ? a.sort - b.sort : a.id - b.id
}
