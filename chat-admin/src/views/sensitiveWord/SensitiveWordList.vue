<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button v-if="can(PERM.OpsSensitiveWordImport)" @click="importVisible = true">
              批量导入
            </el-button>
            <el-button
              v-if="can(PERM.OpsSensitiveWordCreate)"
              type="primary"
              @click="formRef?.openCreate()"
            >
              新增
            </el-button>
          </div>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        show-icon
        class="tip"
        title="启用的屏蔽词保存后立即在所有服务实例生效。匹配规则为包含匹配，忽略大小写、空格与常见分隔符；命中即拒绝发布。"
      />

      <SensitiveWordFilter
        v-model:word="query.word"
        v-model:category="query.category"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />

      <div v-if="selected.length" class="batch-bar">
        <span>已选 {{ selected.length }} 项</span>
        <template v-if="can(PERM.OpsSensitiveWordStatus)">
          <el-button size="small" @click="handleBatchStatus(1)">启用</el-button>
          <el-button size="small" @click="handleBatchStatus(0)">停用</el-button>
        </template>
        <el-button
          v-if="can(PERM.OpsSensitiveWordDelete)"
          size="small"
          type="danger"
          plain
          @click="handleBatchDelete"
        >
          删除
        </el-button>
        <el-button size="small" link @click="tableRef?.clearSelection()">取消选择</el-button>
      </div>

      <SensitiveWordTable
        ref="tableRef"
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        :selectable="canBatch"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @selection-change="(rows) => (selected = rows)"
        @edit="(row) => formRef?.openEdit(row)"
        @toggle-status="handleToggleStatus"
        @delete="handleDelete"
        @page-change="fetchList"
      />
    </el-card>

    <SensitiveWordFormDialog ref="formRef" @success="fetchList" />
    <SensitiveWordImportDialog v-model="importVisible" @success="handleSearch" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  batchDeleteSensitiveWords,
  batchUpdateSensitiveWordStatus,
  deleteSensitiveWord,
  getSensitiveWordList,
  updateSensitiveWordStatus,
  type SensitiveWordItem,
} from '@/api/sensitiveWord'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { DEFAULT_PAGE_SIZE } from '@/constants'
import { PERM } from '@/constants/permission'
import { confirmAndRun } from '@/utils/confirm'
import SensitiveWordFilter from './SensitiveWordFilter.vue'
import SensitiveWordFormDialog from './SensitiveWordFormDialog.vue'
import SensitiveWordImportDialog from './SensitiveWordImportDialog.vue'
import SensitiveWordTable from './SensitiveWordTable.vue'

const { can } = usePermission()
const route = useRoute()
const formRef = ref<InstanceType<typeof SensitiveWordFormDialog>>()
const tableRef = ref<InstanceType<typeof SensitiveWordTable>>()
const importVisible = ref(false)
const selected = ref<SensitiveWordItem[]>([])

const canBatch = computed(
  () => can(PERM.OpsSensitiveWordStatus) || can(PERM.OpsSensitiveWordDelete),
)

const { loading, error, list, total, query, fetchList, handleSearch, handleReset, afterRemove } =
  useListQuery({
    fetch: (q) => getSensitiveWordList(q),
    defaults: {
      page: 1,
      pageSize: DEFAULT_PAGE_SIZE,
      word: '',
      category: '',
      status: undefined as number | undefined,
    },
    schema: { word: 'string', category: 'string', status: 'number' },
    errorMessage: '获取屏蔽词失败',
  })

function preview(rows: SensitiveWordItem[]) {
  const words = rows.slice(0, 5).map((r) => `「${r.word}」`)
  return rows.length > 5 ? `${words.join('、')} 等 ${rows.length} 个词` : words.join('、')
}

async function handleToggleStatus(row: SensitiveWordItem) {
  const next = row.status === 1 ? 0 : 1
  const message =
    next === 1
      ? `确定启用「${row.word}」吗？启用后包含该词的内容将无法发布。`
      : `确定停用「${row.word}」吗？停用后不再拦截该词。`
  await confirmAndRun(
    message,
    async () => {
      await updateSensitiveWordStatus(row.id, next)
      await fetchList()
    },
    next === 1 ? '已启用' : '已停用',
  )
}

async function handleDelete(row: SensitiveWordItem) {
  await confirmAndRun(
    `确定删除屏蔽词「${row.word}」吗？删除后不再拦截该词，且不可恢复。`,
    async () => {
      await deleteSensitiveWord(row.id)
      await afterRemove()
    },
    '已删除',
  )
}

async function handleBatchStatus(status: number) {
  const rows = [...selected.value]
  const action = status === 1 ? '启用' : '停用'
  await confirmAndRun(
    `确定${action}${preview(rows)}吗？`,
    async () => {
      await batchUpdateSensitiveWordStatus(
        rows.map((r) => r.id),
        status,
      )
      tableRef.value?.clearSelection()
      await fetchList()
    },
    `已${action} ${rows.length} 个`,
  )
}

async function handleBatchDelete() {
  const rows = [...selected.value]
  await confirmAndRun(
    `确定删除${preview(rows)}吗？删除后不再拦截这些词，且不可恢复。`,
    async () => {
      await batchDeleteSensitiveWords(rows.map((r) => r.id))
      tableRef.value?.clearSelection()
      await afterRemove(rows.length)
    },
    `已删除 ${rows.length} 个`,
  )
}
</script>

<style scoped>
.tip {
  margin-bottom: 16px;
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  font-size: 13px;
}

.batch-bar .el-button + .el-button {
  margin-left: 0;
}
</style>
