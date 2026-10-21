<template>
  <el-dialog
    v-model="visible"
    title="角色设置"
    width="500px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <p>员工：{{ employee?.name || employee?.mobile }}</p>
    <el-checkbox-group v-model="checkedRoleIds">
      <el-checkbox
        v-for="r in allRoles"
        :key="r.id"
        :label="r.id"
        :disabled="isSuperAdminRole(r) && !operatorIsSuperAdmin"
      >
        {{ r.name }}
      </el-checkbox>
    </el-checkbox-group>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { assignEmployeeRoles, getEmployeeRoles, type EmployeeItem } from '@/api/employee'
import { getRoleList, type RoleItem } from '@/api/role'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ employee: EmployeeItem | null }>()
const emit = defineEmits<{ success: [] }>()

const employeeStore = useEmployeeStore()
const saving = ref(false)
const allRoles = ref<RoleItem[]>([])
const checkedRoleIds = ref<number[]>([])
const operatorIsSuperAdmin = ref(false)

function isSuperAdminRole(role: Pick<RoleItem, 'code'>) {
  return role.code === 'super_admin'
}

watch(visible, async (v) => {
  if (!v || !props.employee) return
  try {
    const operatorId = employeeStore.employeeInfo?.id
    const [roles, employeeRoles, operatorRoles] = await Promise.all([
      getRoleList({ page: 1, pageSize: 200 }),
      getEmployeeRoles(props.employee.id),
      operatorId ? getEmployeeRoles(operatorId) : Promise.resolve({ data: [] as RoleItem[] }),
    ])
    allRoles.value = roles.data.list
    checkedRoleIds.value = employeeRoles.data.map((r) => r.id)
    operatorIsSuperAdmin.value = operatorRoles.data.some((r) => r.code === 'super_admin')

    if (!operatorIsSuperAdmin.value) {
      const superAdminRoleID = allRoles.value.find((r) => isSuperAdminRole(r))?.id
      if (superAdminRoleID) {
        checkedRoleIds.value = checkedRoleIds.value.filter((id) => id !== superAdminRoleID)
      }
    }
  } catch (e) {
    showApiError(e, '获取角色失败')
    visible.value = false
  }
})

async function handleSubmit() {
  if (!props.employee) return
  const assigningSuperAdmin = allRoles.value.some(
    (r) => isSuperAdminRole(r) && checkedRoleIds.value.includes(r.id),
  )
  if (assigningSuperAdmin && !operatorIsSuperAdmin.value) {
    ElMessage.warning('仅超级管理员可分配超级管理员角色')
    return
  }
  saving.value = true
  try {
    await assignEmployeeRoles(props.employee.id, checkedRoleIds.value)
    ElMessage.success('角色设置成功')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '操作失败')
  } finally {
    saving.value = false
  }
}
</script>
