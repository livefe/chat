import { useRouter } from 'vue-router'

import type { PermissionCode } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { pushMenuRoute } from '@/utils/navigation'

export function useDrawerNav() {
  const router = useRouter()
  const employeeStore = useEmployeeStore()
  const drawer = useUserDrawerStore()

  function go(code: PermissionCode, query: Record<string, string>) {
    drawer.close()
    pushMenuRoute(router, employeeStore.permissions, code, query)
  }

  return { go }
}
