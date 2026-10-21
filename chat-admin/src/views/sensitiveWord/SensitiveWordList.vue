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
        title="修改屏蔽词后需重启服务才会生效。匹配规则为包含匹配（会忽略空格与常见分隔符）。"
      />

      <SensitiveWordFilter
        v-model:word="query.word"
        v-model:category="query.category"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />

      <SensitiveWordTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @edit="(row) => formRef?.openEdit(row)"
        @toggle-status="handleToggleStatus"
        @delete="handleDelete"
        @page-change="fetchList"
      />
    </el-card>

    <SensitiveWordFormDialog ref="formRef" @success="fetchList" />
    <SensitiveWordImportDialog v-model="importVisible" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  deleteSensitiveWord,
  getSensitiveWordList,
  updateSensitiveWordStatus,
  type SensitiveWordItem,
} from '@/api/sensitiveWord'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { confirmAndRun } from '@/utils/confirm'
import SensitiveWordFilter from './SensitiveWordFilter.vue'
import SensitiveWordFormDialog from './SensitiveWordFormDialog.vue'
import SensitiveWordImportDialog from './SensitiveWordImportDialog.vue'
import SensitiveWordTable from './SensitiveWordTable.vue'

const { can } = usePermission()
const route = useRoute()
const formRef = ref<InstanceType<typeof SensitiveWordFormDialog>>()
const importVisible = ref(false)

const { loading, list, total, query, fetchList, handleSearch, handleReset } = usePaginatedList(
  getSensitiveWordList,
  { page: 1, pageSize: 20, word: '', category: '', status: undefined as number | undefined },
)

async function handleToggleStatus(row: SensitiveWordItem) {
  const next = row.status === 1 ? 0 : 1
  await confirmAndRun(`确定要${next === 1 ? '启用' : '停用'}「${row.word}」吗？`, async () => {
    await updateSensitiveWordStatus(row.id, next)
    await fetchList()
  })
}

async function handleDelete(row: SensitiveWordItem) {
  await confirmAndRun(
    `确定删除屏蔽词「${row.word}」吗？`,
    async () => {
      await deleteSensitiveWord(row.id)
      await fetchList()
    },
    '已删除',
  )
}

onMounted(fetchList)
</script>

<style scoped>
.tip {
  margin-bottom: 16px;
}
</style>
