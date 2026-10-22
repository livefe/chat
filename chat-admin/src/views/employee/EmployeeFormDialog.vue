<template>
  <el-dialog
    v-model="visible"
    :title="form.id ? '编辑员工' : '新增员工'"
    width="520px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
      <el-form-item label="手机号" prop="mobile">
        <el-input v-model="form.mobile" placeholder="请输入手机号" maxlength="11" />
      </el-form-item>
      <el-form-item label="姓名" prop="name">
        <el-input v-model="form.name" placeholder="请输入姓名" maxlength="32" show-word-limit />
      </el-form-item>
      <el-form-item label="工号">
        <el-input v-model="form.employeeNo" placeholder="请输入工号" maxlength="32" />
      </el-form-item>
      <el-form-item label="部门" prop="departmentId">
        <el-tree-select
          v-model="form.departmentId"
          :data="departmentOptions"
          check-strictly
          filterable
          :render-after-expand="false"
          placeholder="请选择部门"
          style="width: 100%"
          node-key="id"
          :props="{ label: 'name', children: 'children' }"
        />
      </el-form-item>
      <el-form-item label="备注">
        <el-input
          v-model="form.remark"
          type="textarea"
          placeholder="备注"
          maxlength="255"
          show-word-limit
        />
      </el-form-item>
      <el-form-item v-if="!form.id" label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">启用</el-radio>
          <el-radio :value="0">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { createEmployee, updateEmployee } from '@/api/employee'
import type { DepartmentOption } from '@/api/department'
import { MOBILE_RULE } from '@/constants'
import { showApiError } from '@/utils/apiError'

export interface EmployeeFormModel {
  id: number
  mobile: string
  name: string
  employeeNo: string
  departmentId?: number
  remark: string
  status: number
}

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{
  model: EmployeeFormModel
  departmentOptions: DepartmentOption[]
}>()
const emit = defineEmits<{ success: [] }>()

const formRef = ref<FormInstance>()
const form = reactive<EmployeeFormModel>({
  id: 0,
  mobile: '',
  name: '',
  employeeNo: '',
  departmentId: undefined,
  remark: '',
  status: 1,
})
const saving = ref(false)
const rules = {
  mobile: [{ required: true, message: '请输入手机号', trigger: 'blur' }, MOBILE_RULE],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { max: 32, message: '姓名最多 32 个字符', trigger: 'blur' },
  ],
  departmentId: [{ required: true, message: '请选择部门', trigger: 'change' }],
}

watch(visible, (v) => {
  if (v) Object.assign(form, props.model)
})

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  const departmentId = form.departmentId
  if (departmentId === undefined) return
  saving.value = true
  try {
    const payload = {
      mobile: form.mobile,
      name: form.name,
      employeeNo: form.employeeNo,
      departmentId,
      remark: form.remark,
    }
    if (form.id) await updateEmployee({ id: form.id, ...payload })
    else await createEmployee({ ...payload, status: form.status })
    ElMessage.success(form.id ? '已保存' : '已新增')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}
</script>
