<template>
  <el-table :data="list" v-loading="loading" @selection-change="onSelectionChange">
    <el-table-column
      v-if="selectable"
      type="selection"
      width="48"
      :selectable="(row: AdCampaignItem) => row.status === AD_STATUS.Active"
    />
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column prop="advertiserName" label="广告主" min-width="110" show-overflow-tooltip />
    <el-table-column label="标题" min-width="160" show-overflow-tooltip>
      <template #default="{ row }">
        {{ row.title }}
        <div v-if="row.name && row.name !== row.title" class="sub">{{ row.name }}</div>
      </template>
    </el-table-column>
    <el-table-column label="封面" width="72">
      <template #default="{ row }">
        <el-image
          v-if="row.imageUrl"
          :src="row.imageUrl"
          :preview-src-list="[row.imageUrl]"
          fit="cover"
          class="thumb-sm"
          preview-teleported
        />
        <span v-else>-</span>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tooltip
          v-if="row.status === AD_STATUS.Rejected && row.rejectRemark"
          :content="`驳回原因：${row.rejectRemark}`"
          placement="top"
        >
          <el-tag :type="statusTagType(row.status)" size="small" class="tag-help">
            {{ row.statusText }}
          </el-tag>
        </el-tooltip>
        <el-tag v-else :type="statusTagType(row.status)" size="small">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="priority" label="优先级" width="80" />
    <el-table-column label="今日数据" width="150">
      <template #default="{ row }">
        <div class="stat-cell">
          <span>曝光 {{ row.todayImpressions ?? 0 }}</span>
          <span>点击 {{ row.todayClicks ?? 0 }}</span>
          <span>CTR {{ formatAdCtr(row.todayCtr) }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="排期" width="170">
      <template #default="{ row }">
        <div class="stat-cell">
          <span>{{ formatUnixSeconds(row.startAt) }}</span>
          <span>{{ formatUnixSeconds(row.endAt) }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="标记" width="120">
      <template #default="{ row }">
        <div class="tag-stack">
          <el-tag v-if="row.landingInvalid" type="danger" size="small">落地失效</el-tag>
          <el-tag v-if="row.lowCtrWarning" type="warning" size="small">低 CTR</el-tag>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="280" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="canEdit(row as AdCampaignItem)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as AdCampaignItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="canSubmitAudit(row as AdCampaignItem)"
            size="small"
            link
            type="primary"
            @click="emit('status', row as AdCampaignItem, AD_STATUS.Pending)"
          >
            提交审核
          </el-button>
          <el-button
            v-if="canAudit(row as AdCampaignItem)"
            size="small"
            link
            type="warning"
            @click="emit('audit', row as AdCampaignItem)"
          >
            审核
          </el-button>
          <el-button
            v-if="canChangeStatus && row.status === AD_STATUS.Paused"
            size="small"
            link
            type="success"
            @click="emit('status', row as AdCampaignItem, AD_STATUS.Active)"
          >
            恢复投放
          </el-button>
          <el-button
            v-if="canChangeStatus && row.status === AD_STATUS.Active"
            size="small"
            link
            type="warning"
            @click="emit('status', row as AdCampaignItem, AD_STATUS.Paused)"
          >
            暂停
          </el-button>
          <el-button
            v-if="canEnd(row as AdCampaignItem)"
            size="small"
            link
            type="info"
            @click="emit('status', row as AdCampaignItem, AD_STATUS.Ended)"
          >
            结束
          </el-button>
          <el-button
            v-if="can(PERM.AdCampaignCopy)"
            size="small"
            link
            @click="emit('copy', row as AdCampaignItem)"
          >
            复制
          </el-button>
          <el-button
            v-if="can(PERM.AdCampaignStats)"
            size="small"
            link
            @click="emit('stats', row as AdCampaignItem)"
          >
            日报
          </el-button>
        </div>
      </template>
    </el-table-column>
    <template #empty>
      <TableEmpty :error="error" text="暂无广告计划" @retry="emit('page-change')" />
    </template>
  </el-table>

  <div class="pagination">
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="PAGE_SIZES"
      :layout="PAGINATION_LAYOUT"
      @size-change="emit('page-change')"
      @current-change="emit('page-change')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { formatUnixSeconds } from '@/utils/datetime'
import { type AdCampaignItem } from '@/api/ad'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { AD_STATUS, formatAdCtr } from '@/constants/ad'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: AdCampaignItem[]
  loading: boolean
  error: boolean
  total: number
  selectable: boolean
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  'selection-change': [AdCampaignItem[]]
  edit: [AdCampaignItem]
  status: [AdCampaignItem, number]
  audit: [AdCampaignItem]
  copy: [AdCampaignItem]
  stats: [AdCampaignItem]
  'page-change': []
}>()

const { can } = usePermission()
const canChangeStatus = computed(() => can(PERM.AdCampaignStatus))

function onSelectionChange(rows: AdCampaignItem[]) {
  emit('selection-change', rows)
}

function statusTagType(status: number) {
  switch (status) {
    case AD_STATUS.Active:
      return 'success'
    case AD_STATUS.Pending:
    case AD_STATUS.Paused:
      return 'warning'
    case AD_STATUS.Rejected:
      return 'danger'
    default:
      return 'info'
  }
}

function canEdit(row: AdCampaignItem) {
  return (
    can(PERM.AdCampaignUpdate) &&
    (row.status === AD_STATUS.Draft ||
      row.status === AD_STATUS.Rejected ||
      row.status === AD_STATUS.Paused)
  )
}

function canSubmitAudit(row: AdCampaignItem) {
  return (
    canChangeStatus.value && (row.status === AD_STATUS.Draft || row.status === AD_STATUS.Rejected)
  )
}

function canAudit(row: AdCampaignItem) {
  return can(PERM.AdCampaignAudit) && row.status === AD_STATUS.Pending
}

function canEnd(row: AdCampaignItem) {
  return canChangeStatus.value && row.status !== AD_STATUS.Ended && row.status !== AD_STATUS.Draft
}
</script>

<style scoped>
.stat-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--el-text-color-regular);
}
.tag-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-start;
}
.tag-help {
  cursor: help;
}
</style>
