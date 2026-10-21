<template>
  <el-dialog
    v-model="visible"
    :title="form.id ? '编辑屏蔽词' : '新增屏蔽词'"
    width="480px"
    destroy-on-close
    :close-on-click-modal="false"
    @closed="onClosed"
  >
    <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
      <el-form-item label="屏蔽词" prop="word">
        <el-input v-model="form.word" maxlength="64" show-word-limit placeholder="请输入屏蔽词" />
      </el-form-item>
      <el-form-item label="分类" prop="category">
        <el-select v-model="form.category" style="width: 100%">
          <el-option
            v-for="opt in SENSITIVE_CATEGORY_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="等级" prop="level">
        <el-select v-model="form.level" style="width: 100%">
          <el-option
            v-for="opt in SENSITIVE_LEVEL_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">启用</el-radio>
          <el-radio :value="0">停用</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="备注" prop="remark">
        <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="255" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  createSensitiveWord,
  updateSensitiveWord,
  type SensitiveWordItem,
} from '@/api/sensitiveWord'
import { SENSITIVE_CATEGORY_OPTIONS, SENSITIVE_LEVEL_OPTIONS } from '@/constants/sensitiveWord'
import { showApiError } from '@/utils/apiError'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

function defaultForm() {
  return { id: 0, word: '', category: 'other', level: 1, status: 1, remark: '' }
}

const form = reactive(defaultForm())

const formRules: FormRules = {
  word: [{ required: true, message: '请输入屏蔽词', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  level: [{ required: true, message: '请选择等级', trigger: 'change' }],
}

async function open(data: ReturnType<typeof defaultForm>) {
  Object.assign(form, data)
  visible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

function onClosed() {
  Object.assign(form, defaultForm())
}

function openCreate() {
  void open(defaultForm())
}

function openEdit(row: SensitiveWordItem) {
  void open({
    id: row.id,
    word: row.word,
    category: row.category || 'other',
    level: row.level || 1,
    status: row.status,
    remark: row.remark || '',
  })
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  saving.value = true
  try {
    const payload = {
      word: form.word.trim(),
      category: form.category,
      level: form.level,
      status: form.status,
      remark: form.remark,
    }
    if (form.id) await updateSensitiveWord({ id: form.id, ...payload })
    else await createSensitiveWord(payload)
    ElMessage.success('操作成功，重启服务后生效')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '操作失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ openCreate, openEdit })
</script>
