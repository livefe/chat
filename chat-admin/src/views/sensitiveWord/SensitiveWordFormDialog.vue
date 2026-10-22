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
        <el-input
          v-model="form.word"
          :maxlength="SENSITIVE_WORD_MAX_LEN"
          show-word-limit
          placeholder="请输入屏蔽词"
        />
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
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio v-for="opt in SENSITIVE_STATUS_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </el-radio>
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
import {
  SENSITIVE_CATEGORY_OPTIONS,
  SENSITIVE_STATUS_OPTIONS,
  SENSITIVE_WORD_MAX_LEN,
} from '@/constants/sensitiveWord'
import { showApiError } from '@/utils/apiError'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

function defaultForm() {
  return { id: 0, word: '', category: 'other', status: 1, remark: '' }
}

const form = reactive(defaultForm())

const formRules: FormRules = {
  word: [{ required: true, whitespace: true, message: '请输入屏蔽词', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
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
    status: row.status,
    remark: row.remark || '',
  })
}

async function handleSubmit() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    const payload = {
      word: form.word.trim(),
      category: form.category,
      status: form.status,
      remark: form.remark,
    }
    if (form.id) await updateSensitiveWord({ id: form.id, ...payload })
    else await createSensitiveWord(payload)
    ElMessage.success(form.id ? '已保存，已立即生效' : '已新增，已立即生效')
    visible.value = false
    emit('success')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}

defineExpose({ openCreate, openEdit })
</script>
