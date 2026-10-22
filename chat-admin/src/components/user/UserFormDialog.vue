<template>
  <el-dialog
    v-model="formDialog.visible"
    :title="isEdit ? '编辑用户' : '新增用户'"
    width="500px"
    destroy-on-close
    :close-on-click-modal="false"
    @open="reset"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item :label="isEdit ? '新手机号' : '手机号'" prop="mobile">
        <el-input
          v-model="form.mobile"
          :placeholder="
            isEdit ? `留空则不修改（当前 ${formDialog.model.maskedMobile || '-'}）` : '请输入手机号'
          "
        />
      </el-form-item>
      <el-form-item label="昵称" prop="nickname">
        <el-input v-model="form.nickname" placeholder="请输入昵称" maxlength="32" show-word-limit />
      </el-form-item>
      <el-form-item v-if="!isEdit" label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio :value="UserStatus.Enabled">启用</el-radio>
          <el-radio :value="UserStatus.Disabled">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="formDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'

import { createUser, updateUser } from '@/api/user'
import { useUserActionDialogs } from '@/composables/useUserActions'
import { MOBILE_RULE } from '@/constants'
import { UserStatus } from '@/constants/user'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { showApiError } from '@/utils/apiError'

const { formDialog } = useUserActionDialogs()
const drawer = useUserDrawerStore()

const formRef = ref<FormInstance>()
const saving = ref(false)
const form = reactive({ mobile: '', nickname: '', status: UserStatus.Enabled as number })
const isEdit = computed(() => formDialog.model.id > 0)

const rules = computed<FormRules>(() => ({
  mobile: isEdit.value
    ? [MOBILE_RULE]
    : [{ required: true, message: '请输入手机号', trigger: 'blur' }, MOBILE_RULE],
  nickname: [
    { required: true, message: '请填写昵称', trigger: 'blur' },
    { max: 32, message: '昵称最多 32 个字符', trigger: 'blur' },
  ],
}))

function reset() {
  form.mobile = ''
  form.nickname = formDialog.model.nickname
  form.status = UserStatus.Enabled
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  saving.value = true
  try {
    const nickname = form.nickname.trim()
    if (isEdit.value) {
      await updateUser({ id: formDialog.model.id, mobile: form.mobile, nickname })
    } else {
      await createUser({ mobile: form.mobile, nickname, status: form.status })
    }
    ElMessage.success(isEdit.value ? '已保存' : '已新增')
    formDialog.visible = false
    drawer.refresh()
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}
</script>
