<template>
  <el-dialog
    v-model="visible"
    :title="form.id ? '编辑角色' : '新增角色'"
    width="500px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入角色名称" />
      </el-form-item>
      <el-form-item label="编码" prop="code">
        <el-input v-model="form.code" placeholder="请输入角色编码" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入角色描述"
        />
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="form.status">
          <el-radio :label="1">启用</el-radio>
          <el-radio :label="0">禁用</el-radio>
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
import { createRole, updateRole } from '@/api/role'
import { showApiError } from '@/utils/apiError'

export interface RoleFormModel {
  id: number
  name: string
  code: string
  description: string
  status: number
}

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ model: RoleFormModel }>()
const emit = defineEmits<{ success: [] }>()

const formRef = ref<FormInstance>()
const form = reactive<RoleFormModel>({ id: 0, name: '', code: '', description: '', status: 1 })
const saving = ref(false)
const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
}

watch(visible, (v) => {
  if (v) Object.assign(form, props.model)
})

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  saving.value = true
  try {
    if (form.id) await updateRole(form)
    else await createRole(form)
    ElMessage.success('操作成功')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '操作失败')
  } finally {
    saving.value = false
  }
}
</script>
