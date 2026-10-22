<template>
  <el-dialog
    v-model="visible"
    title="角色设置"
    width="520px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <div v-loading="loading" class="role-dialog">
      <p class="role-dialog__target">员工：{{ employee?.name || employee?.mobile }}</p>
      <el-alert
        v-if="readonly"
        type="warning"
        show-icon
        :closable="false"
        class="role-dialog__tip"
        title="该员工是超级管理员，只有超级管理员可以调整其角色。"
      />
      <el-alert
        v-else-if="disabledAssigned.length"
        type="info"
        show-icon
        :closable="false"
        class="role-dialog__tip"
        :title="`已禁用的角色（${disabledAssigned.map((r) => r.name).join('、')}）不会生效，可取消勾选移除。`"
      />
      <el-checkbox-group v-model="checkedRoleIds" :disabled="readonly">
        <el-checkbox
          v-for="r in options"
          :key="r.id"
          :value="r.id"
          :disabled="isOptionDisabled(r)"
          class="role-dialog__option"
        >
          <span :class="{ 'is-disabled-role': r.status !== 1 }">{{ r.name }}</span>
          <el-tag v-if="r.status !== 1" size="small" type="info" class="role-dialog__tag">
            已禁用
          </el-tag>
          <el-tag v-else-if="isSuperRole(r)" size="small" type="warning" class="role-dialog__tag">
            仅超管可分配
          </el-tag>
        </el-checkbox>
      </el-checkbox-group>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">{{ readonly ? '关闭' : '取消' }}</el-button>
        <el-button
          v-if="!readonly"
          type="primary"
          :loading="saving"
          :disabled="loading"
          @click="handleSubmit"
        >
          确定
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  assignEmployeeRoles,
  getEmployeeRoleOptions,
  type EmployeeItem,
  type EmployeeRoleBrief,
} from '@/api/employee'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'

const SUPER_ADMIN_ROLE_CODE = 'super_admin'

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ employee: EmployeeItem | null }>()
const emit = defineEmits<{ success: [] }>()

const employeeStore = useEmployeeStore()
const loading = ref(false)
const saving = ref(false)
const options = ref<EmployeeRoleBrief[]>([])
const assignedIds = ref<number[]>([])
const checkedRoleIds = ref<number[]>([])
const operatorIsSuper = ref(false)
const targetIsSuper = ref(false)

const readonly = computed(() => targetIsSuper.value && !operatorIsSuper.value)
const disabledAssigned = computed(() =>
  options.value.filter((r) => r.status !== 1 && assignedIds.value.includes(r.id)),
)

function isSuperRole(role: EmployeeRoleBrief) {
  return role.code === SUPER_ADMIN_ROLE_CODE
}

/** 禁用角色只能取消、不能新勾选；超管角色只有超管能动。 */
function isOptionDisabled(role: EmployeeRoleBrief) {
  if (role.status !== 1 && !checkedRoleIds.value.includes(role.id)) return true
  return isSuperRole(role) && !operatorIsSuper.value
}

watch(visible, async (v) => {
  if (!v || !props.employee) return
  loading.value = true
  options.value = []
  checkedRoleIds.value = []
  try {
    const { data } = await getEmployeeRoleOptions(props.employee.id)
    options.value = data.options
    assignedIds.value = data.assignedIds
    checkedRoleIds.value = [...data.assignedIds]
    operatorIsSuper.value = data.operatorIsSuper
    targetIsSuper.value = data.targetIsSuper
  } catch (e) {
    showApiError(e, '获取角色失败')
    visible.value = false
  } finally {
    loading.value = false
  }
})

async function handleSubmit() {
  if (!props.employee) return
  saving.value = true
  try {
    await assignEmployeeRoles(props.employee.id, checkedRoleIds.value)
    ElMessage.success('角色已更新')
    visible.value = false
    emit('success')
    if (props.employee.id === employeeStore.employeeInfo?.id) {
      await employeeStore.loadPermissions()
    }
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.role-dialog {
  min-height: 80px;
}

.role-dialog__target {
  margin: 0 0 12px;
}

.role-dialog__tip {
  margin-bottom: 12px;
}

.role-dialog__option {
  margin-right: 24px;
}

.role-dialog__tag {
  margin-left: 6px;
}

.is-disabled-role {
  color: var(--el-text-color-placeholder);
  text-decoration: line-through;
}
</style>
