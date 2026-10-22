<template>
  <el-dialog
    v-model="visible"
    title="批量导入"
    width="560px"
    destroy-on-close
    :close-on-click-modal="false"
    @closed="reset"
  >
    <template v-if="!result">
      <el-form label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="category" style="width: 100%">
            <el-option
              v-for="opt in SENSITIVE_CATEGORY_OPTIONS"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="词列表">
          <div class="import-input">
            <el-upload
              :show-file-list="false"
              accept=".txt,text/plain"
              :auto-upload="false"
              :on-change="onFileChange"
            >
              <el-button size="small">从 .txt 文件读取</el-button>
            </el-upload>
            <span class="form-tip">每行一个词，最多 {{ SENSITIVE_WORD_MAX_LEN }} 字</span>
          </div>
          <el-input
            v-model="words"
            type="textarea"
            :rows="10"
            placeholder="每行一个词，也可直接粘贴"
          />
          <div class="form-tip">共 {{ lineCount }} 行；导入的词默认启用，导入后立即生效</div>
        </el-form-item>
      </el-form>
    </template>

    <template v-else>
      <el-result
        :icon="result.created > 0 ? 'success' : 'warning'"
        :title="`新增 ${result.created} 个，跳过 ${skippedTotal} 个`"
      />
      <el-collapse v-if="skippedTotal > 0">
        <el-collapse-item
          v-for="group in skippedGroups"
          :key="group.title"
          :title="`${group.title}（${group.words.length}）`"
        >
          <div class="skipped-words">
            <el-tag v-for="(w, i) in group.words" :key="`${w}-${i}`" type="info">{{ w }}</el-tag>
          </div>
        </el-collapse-item>
      </el-collapse>
    </template>

    <template #footer>
      <template v-if="!result">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="handleImport">导入</el-button>
      </template>
      <template v-else>
        <el-button @click="result = null">继续导入</el-button>
        <el-button type="primary" @click="visible = false">完成</el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, type UploadFile } from 'element-plus'
import { importSensitiveWords, type ImportSensitiveWordsResult } from '@/api/sensitiveWord'
import { SENSITIVE_CATEGORY_OPTIONS, SENSITIVE_WORD_MAX_LEN } from '@/constants/sensitiveWord'
import { showApiError } from '@/utils/apiError'

const MAX_FILE_BYTES = 1024 * 1024

const visible = defineModel<boolean>({ required: true })
const emit = defineEmits<{ success: [] }>()

const importing = ref(false)
const words = ref('')
const category = ref('other')
const result = ref<ImportSensitiveWordsResult | null>(null)

const lineCount = computed(() => words.value.split('\n').filter((l) => l.trim()).length)

const skippedGroups = computed(() => {
  const r = result.value
  if (!r) return []
  return [
    { title: '已存在', words: r.existing },
    { title: `超过 ${SENSITIVE_WORD_MAX_LEN} 字`, words: r.tooLong },
    { title: '本次重复', words: r.duplicated },
    { title: '只含空格或分隔符', words: r.invalid },
  ].filter((g) => g.words.length > 0)
})

const skippedTotal = computed(() => skippedGroups.value.reduce((n, g) => n + g.words.length, 0))

async function onFileChange(file: UploadFile) {
  const raw = file.raw
  if (!raw) return
  if (raw.size > MAX_FILE_BYTES) {
    ElMessage.warning('文件不能超过 1MB')
    return
  }
  const text = await raw.text()
  words.value = words.value.trim() ? `${words.value.trimEnd()}\n${text}` : text
}

function reset() {
  words.value = ''
  category.value = 'other'
  result.value = null
}

async function handleImport() {
  if (!lineCount.value) {
    ElMessage.warning('请输入或选择要导入的词')
    return
  }
  importing.value = true
  try {
    const { data } = await importSensitiveWords({
      words: words.value,
      category: category.value,
    })
    result.value = data
    words.value = ''
    if (data.created > 0) emit('success')
  } catch (e) {
    showApiError(e, '导入失败')
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.import-input {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.skipped-words {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
</style>
