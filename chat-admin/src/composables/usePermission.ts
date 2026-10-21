import type { PermissionCode } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'

export function usePermission() {
  const employeeStore = useEmployeeStore()

  function can(code: PermissionCode) {
    return employeeStore.hasPermission(code)
  }

  return { can }
}
