<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button
              v-if="can(PERM.AdCampaignPause)"
              :disabled="!selectedIds.length"
              @click="handleBatchPause"
            >
              批量暂停
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
        @search="handleSearch"
        @reset="handleReset"
      />
      <AdCampaignTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @selection-change="onSelectionChange"
        @edit="handleEdit"
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
      @approve="submitAudit(true)"
      @reject="submitAudit(false, $event)"
    />

    <el-drawer v-model="statsVisible" :title="statsTitle" size="520px" destroy-on-close>
      <el-table :data="statsList" v-loading="statsLoading" size="small">
        <el-table-column prop="statDate" label="日期" width="120" />
        <el-table-column prop="impressions" label="曝光" />
        <el-table-column prop="clicks" label="点击" />
        <el-table-column label="CTR">
          <template #default="{ row }">{{ formatCtr(row.ctr) }}</template>
        </el-table-column>
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  batchPauseAdCampaigns,
  copyAdCampaign,
  getAdCampaignDailyStats,
  getAdCampaignList,
  updateAdCampaignStatus,
  type AdCampaignItem,
  type AdDailyStatItem,
} from '@/api/ad'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { confirmAndRun } from '@/utils/confirm'
import { showApiError } from '@/utils/apiError'
import AdCampaignFilter from './AdCampaignFilter.vue'
import AdCampaignTable from './AdCampaignTable.vue'
import AdCampaignFormDialog from './AdCampaignFormDialog.vue'
import AdCampaignAuditDialog from './AdCampaignAuditDialog.vue'
import { AD_STATUS } from '@/constants/ad'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const { can } = usePermission()
const route = useRoute()
const { loading, list, total, query, fetchList, handleSearch, handleReset } = usePaginatedList(
  getAdCampaignList,
  { page: 1, pageSize: 20, keyword: '', status: undefined as number | undefined },
)

const selectedIds = ref<number[]>([])
const formDialog = ref<InstanceType<typeof AdCampaignFormDialog> | null>(null)
const auditVisible = ref(false)
const auditCampaignId = ref(0)
const auditSubmitting = ref(false)

function onSelectionChange(rows: AdCampaignItem[]) {
  selectedIds.value = rows.map((r) => r.id)
}

function formatCtr(ctr?: number) {
  if (ctr == null || Number.isNaN(ctr)) return '0%'
  return `${(ctr * 100).toFixed(2)}%`
}

function handleAdd() {
  formDialog.value?.openCreate()
}

function handleEdit(row: AdCampaignItem) {
  void formDialog.value?.openEdit(row)
}

async function handleStatus(row: AdCampaignItem, status: number, label: string) {
  await confirmAndRun(`确定对「${row.title}」执行${label}？`, async () => {
    await updateAdCampaignStatus({ id: row.id, status })
    await fetchList()
  })
}

function openAudit(row: AdCampaignItem) {
  auditCampaignId.value = row.id
  auditVisible.value = true
}

async function submitAudit(approve: boolean, rejectRemark = '') {
  if (!auditCampaignId.value) return
  auditSubmitting.value = true
  try {
    await updateAdCampaignStatus({
      id: auditCampaignId.value,
      status: approve ? AD_STATUS.Active : AD_STATUS.Rejected,
      rejectRemark: approve ? undefined : rejectRemark,
    })
    ElMessage.success(approve ? '已通过' : '已驳回')
    auditVisible.value = false
    await fetchList()
  } catch (e) {
    showApiError(e, approve ? '审核通过失败' : '驳回失败')
  } finally {
    auditSubmitting.value = false
  }
}

async function handleCopy(row: AdCampaignItem) {
  await confirmAndRun(
    `确定复制「${row.title}」？`,
    async () => {
      await copyAdCampaign(row.id)
      await fetchList()
    },
    '复制成功',
  )
}

async function handleBatchPause() {
  if (!selectedIds.value.length) return
  await confirmAndRun(
    `确定暂停选中的 ${selectedIds.value.length} 条计划？`,
    async () => {
      await batchPauseAdCampaigns(selectedIds.value)
      selectedIds.value = []
      await fetchList()
    },
    '已暂停',
  )
}

const statsVisible = ref(false)
const statsLoading = ref(false)
const statsList = ref<AdDailyStatItem[]>([])
const statsTitle = ref('每日数据')

async function openStats(row: AdCampaignItem) {
  statsTitle.value = `每日数据 · ${row.title} (#${row.id})`
  statsVisible.value = true
  statsLoading.value = true
  statsList.value = []
  try {
    const { data } = await getAdCampaignDailyStats({ campaignId: row.id })
    statsList.value = data.list || []
  } catch (e) {
    showApiError(e, '加载日报失败')
  } finally {
    statsLoading.value = false
  }
}

onMounted(fetchList)
</script>
