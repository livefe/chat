<template>
  <el-dialog
    v-model="visible"
    title="批量导入"
    width="520px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-form label-width="80px">
      <el-form-item label="分类">
        <el-select v-model="form.category" style="width: 100%">
          <el-option
            v-for="opt in SENSITIVE_CATEGORY_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="等级">
        <el-select v-model="form.level" style="width: 100%">
          <el-option
            v-for="opt in SENSITIVE_LEVEL_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="词列表">
        <el-input
          v-model="form.words"
          type="textarea"
          :rows="10"
          placeholder="每行一个词，# 开头为注释"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="importing" @click="handleImport">导入</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { importSensitiveWords } from '@/api/sensitiveWord'
import { SENSITIVE_CATEGORY_OPTIONS, SENSITIVE_LEVEL_OPTIONS } from '@/constants/sensitiveWord'
import { showApiError } from '@/utils/apiError'

const visible = defineModel<boolean>({ required: true })
const emit = defineEmits<{ success: [] }>()

const importing = ref(false)
const form = reactive({
  words: '',
  category: 'other',
  level: 1,
})

async function handleImport() {
  if (!form.words.trim()) {
    ElMessage.warning('请输入要导入的词')
    return
  }
  importing.value = true
  try {
    const { data } = await importSensitiveWords({
      words: form.words,
      category: form.category,
      level: form.level,
    })
    ElMessage.success(`导入完成：新增 ${data.created}，跳过 ${data.skipped}。重启服务后生效`)
    visible.value = false
    form.words = ''
    emit('success')
  } catch (e) {
    showApiError(e, '导入失败')
  } finally {
    importing.value = false
  }
}
</script>
