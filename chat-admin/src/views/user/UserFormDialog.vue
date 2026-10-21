<template>
  <el-dialog
    v-model="visible"
    :title="form.id ? '编辑用户' : '新增用户'"
    width="500px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
      <el-form-item label="手机号" prop="mobile">
        <el-input v-model="form.mobile" placeholder="请输入手机号" />
      </el-form-item>
      <el-form-item label="昵称" prop="nickname">
        <el-input v-model="form.nickname" placeholder="请输入昵称" maxlength="32" show-word-limit />
      </el-form-item>
      <el-form-item label="状态" prop="status">
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
import { showApiError } from '@/utils/apiError'
import { MOBILE_RULE } from '@/constants'
import { createUser, updateUser } from '@/api/user'

export interface UserFormModel {
  id: number
  mobile: string
  nickname: string
  status: number
}

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ model: UserFormModel }>()
const emit = defineEmits<{ success: [] }>()

const formRef = ref<FormInstance>()
const form = reactive<UserFormModel>({ id: 0, mobile: '', nickname: '', status: 1 })
const saving = ref(false)
const rules = {
  mobile: [{ required: true, message: '请输入手机号', trigger: 'blur' }, MOBILE_RULE],
  nickname: [
    { required: true, message: '请填写昵称', trigger: 'blur' },
    { max: 32, message: '昵称最多 32 个字符', trigger: 'blur' },
  ],
}

watch(visible, (v) => {
  if (v) Object.assign(form, props.model)
})

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  saving.value = true
  try {
    const payload = {
      mobile: form.mobile,
      nickname: form.nickname.trim(),
      status: form.status,
    }
    if (form.id) await updateUser({ id: form.id, ...payload })
    else await createUser(payload)
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
