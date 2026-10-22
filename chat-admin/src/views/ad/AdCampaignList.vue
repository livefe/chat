<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button v-if="canBatchPause" :disabled="!selected.length" @click="handleBatchPause">
              批量暂停{{ selected.length ? `（${selected.length}）` : '' }}
            </el-button>
            <el-button v-if="can(PERM.AdCampaignUpdate)" type="primary" @click="handleAdd">
              新建计划
            </el-button>
          </div>
        </div>
      </template>

      <AdCampaignFilter
        v-model:keyword="query.keyword"
        v-model:status="query.status"
        v-model:advertiser-name="query.advertiserName"
        v-model:start-date="query.startDate"
        v-model:end-date="query.endDate"
        @search="handleSearch"
        @reset="handleReset"
      />
      <div v-if="canBatchPause" class="batch-hint">仅「投放中」的计划可勾选批量暂停</div>
      <AdCampaignTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        :selectable="canBatchPause"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @selection-change="(rows) => (selected = rows)"
        @edit="(row) => formDialog?.openEdit(row.id)"
        @status="handleStatus"
        @audit="openAudit"
        @copy="handleCopy"
        @stats="openStats"
        @page-change="fetchList"
      />
    </el-card>

    <AdCampaignFormDialog ref="formDialog" @success="fetchList" />

    <AdCampaignAuditDialog
      :visible="auditVisible"
      :campaign-id="auditCampaignId"
      :submitting="auditSubmitting"
      @update:visible="auditVisible = $event"
      @approve="submitAudit('approve')"
      @reject="submitAudit('reject', $event)"
    />

    <AdCampaignStatsDrawer
      v-model="statsVisible"
      :campaign-id="statsCampaign.id"
      :campaign-title="statsCampaign.title"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  auditAdCampaign,
  batchPauseAdCampaigns,
  copyAdCampaign,
  getAdCampaignList,
  updateAdCampaignStatus,
  type AdCampaignItem,
} from '@/api/ad'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { AD_STATUS } from '@/constants/ad'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'
import AdCampaignFilter from './AdCampaignFilter.vue'
import AdCampaignTable from './AdCampaignTable.vue'
import AdCampaignFormDialog from './AdCampaignFormDialog.vue'
import AdCampaignAuditDialog from './AdCampaignAuditDialog.vue'
import AdCampaignStatsDrawer from './AdCampaignStatsDrawer.vue'

const { can } = usePermission()
const route = useRoute()
const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getAdCampaignList(q),
  defaults: {
    page: 1,
    pageSize: 20,
    keyword: '',
    status: undefined as number | undefined,
    advertiserName: '',
    startDate: '',
    endDate: '',
  },
  schema: {
    keyword: 'string',
    status: 'number',
    advertiserName: 'string',
    startDate: 'string',
    endDate: 'string',
  },
  errorMessage: '获取广告计划失败',
})

const canBatchPause = computed(() => can(PERM.AdCampaignPause))
const selected = ref<AdCampaignItem[]>([])
const formDialog = ref<InstanceType<typeof AdCampaignFormDialog> | null>(null)
const auditVisible = ref(false)
const auditCampaignId = ref(0)
const auditSubmitting = ref(false)
const statsVisible = ref(false)
const statsCampaign = ref({ id: 0, title: '' })

const STATUS_ACTIONS: Record<number, { label: string; message: (title: string) => string }> = {
  [AD_STATUS.Pending]: {
    label: '提交审核',
    message: (t) => `确定将「${t}」提交审核？审核期间不可编辑。`,
  },
  [AD_STATUS.Active]: {
    label: '恢复投放',
    message: (t) => `确定恢复投放「${t}」？恢复后立即按排期展示。`,
  },
  [AD_STATUS.Paused]: {
    label: '暂停',
    message: (t) => `确定暂停「${t}」？暂停后立即停止展示，可随时恢复投放。`,
  },
  [AD_STATUS.Ended]: {
    label: '结束',
    message: (t) => `确定结束「${t}」？结束后不可恢复，如需再次投放请复制该计划。`,
  },
}

function handleAdd() {
  formDialog.value?.openCreate()
}

async function handleStatus(row: AdCampaignItem, status: number) {
  const action = STATUS_ACTIONS[status]
  if (!action) return
  try {
    await ElMessageBox.confirm(action.message(row.title), action.label, {
      type: 'warning',
      confirmButtonText: action.label,
    })
  } catch {
    return
  }
  try {
    await updateAdCampaignStatus({ id: row.id, status })
    ElMessage.success(`已${action.label}`)
    await fetchList()
  } catch (e) {
    showApiError(e, `${action.label}失败`)
  }
}

function openAudit(row: AdCampaignItem) {
  auditCampaignId.value = row.id
  auditVisible.value = true
}

async function submitAudit(action: 'approve' | 'reject', rejectRemark = '') {
  if (!auditCampaignId.value) return
  auditSubmitting.value = true
  try {
    await auditAdCampaign({ id: auditCampaignId.value, action, rejectRemark })
    ElMessage.success(action === 'approve' ? '已通过并上架' : '已驳回')
    auditVisible.value = false
    await fetchList()
  } catch (e) {
    showApiError(e, action === 'approve' ? '审核通过失败' : '驳回失败')
  } finally {
    auditSubmitting.value = false
  }
}

async function handleCopy(row: AdCampaignItem) {
  try {
    const { data } = await copyAdCampaign(row.id)
    ElMessage.success('已复制为草稿，请检查后保存')
    await fetchList()
    if (can(PERM.AdCampaignUpdate)) await formDialog.value?.openEdit(data.id)
  } catch (e) {
    showApiError(e, '复制失败')
  }
}

async function handleBatchPause() {
  const rows = selected.value.filter((r) => r.status === AD_STATUS.Active)
  if (!rows.length) return
  try {
    await ElMessageBox.confirm(
      `确定暂停选中的 ${rows.length} 条投放中计划？暂停后立即停止展示。`,
      '批量暂停',
      { type: 'warning', confirmButtonText: '暂停' },
    )
  } catch {
    return
  }
  try {
    const { data } = await batchPauseAdCampaigns(rows.map((r) => r.id))
    const skipped = data.skipped ? `，${data.skipped} 条因状态已变化未处理` : ''
    ElMessage.success(`已暂停 ${data.paused} 条${skipped}`)
    selected.value = []
    await fetchList()
  } catch (e) {
    showApiError(e, '批量暂停失败')
  }
}

function openStats(row: AdCampaignItem) {
  statsCampaign.value = { id: row.id, title: row.title }
  statsVisible.value = true
}
</script>

<style scoped>
.batch-hint {
  margin: -4px 0 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
