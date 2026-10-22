import { PermissionType } from '@/constants/permission'

export const permissionTypeOptions = [
  { label: '菜单', value: PermissionType.Menu },
  { label: '按钮', value: PermissionType.Button },
  { label: '接口', value: PermissionType.API },
]

export const PERMISSION_MODEL_HINT =
  '菜单 = 可查看页面及其查询接口；按钮 = 可执行操作；接口跟随同编码的菜单或按钮生效'

export const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE']

export function permissionTypeLabel(type: number): string {
  return permissionTypeOptions.find((o) => o.value === type)?.label ?? '未知'
}

export function permissionTypeTag(type: number): 'primary' | 'warning' | 'info' {
  switch (type) {
    case PermissionType.Menu:
      return 'primary'
    case PermissionType.Button:
      return 'warning'
    default:
      return 'info'
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
