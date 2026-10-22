<template>
  <div class="user-support-tab">
    <el-table :data="list" v-loading="loading" size="small">
      <el-table-column label="会话" width="96">
        <template #default="{ row }"><CopyText :text="row.id" /></template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag
            size="small"
            :type="row.status === SupportSessionStatus.Closed ? 'info' : 'success'"
          >
            {{ row.statusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="坐席" width="96" show-overflow-tooltip>
        <template #default="{ row }">{{ row.assigneeName || '未接入' }}</template>
      </el-table-column>
      <el-table-column prop="createdAt" label="发起时间" width="160" />
      <el-table-column label="结案" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <template v-if="row.closedAt">
            {{ row.closedAt }}
            <span v-if="row.closeReasonText" class="sub"> · {{ row.closeReasonText }}</span>
            <div v-if="row.closeNote" class="sub">{{ row.closeNote }}</div>
          </template>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="lastMessage" label="最后消息" min-width="140" show-overflow-tooltip />
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-button
            size="small"
            link
            type="primary"
            @click="openTranscript(row as SupportUserSession)"
          >
            查看记录
          </el-button>
        </template>
      </el-table-column>
      <template #empty>
        <TableEmpty :error="error" text="暂无客服咨询记录" @retry="fetchList" />
      </template>
    </el-table>
    <div v-if="total > query.pageSize" class="user-support-tab__pager">
      <el-pagination
        v-model:current-page="query.page"
        :page-size="query.pageSize"
        :total="total"
        layout="total, prev, pager, next"
        small
        @current-change="fetchList"
      />
    </div>

    <SupportTranscriptDialog
      v-model="transcriptVisible"
      :conversation-id="transcriptId"
      :title="transcriptTitle"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getUserSupportSessions, type SupportUserSession } from '@/api/support'
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { SupportSessionStatus } from '@/constants/support'
import SupportTranscriptDialog from '@/views/support/SupportTranscriptDialog.vue'

const props = defineProps<{ userId: number }>()

const { query, list, total, loading, error, fetchList, handleSearch } = useListQuery({
  fetch: (q) => getUserSupportSessions({ ...q, userId: props.userId }),
  defaults: { page: 1, pageSize: 10 },
  errorMessage: '获取客服记录失败',
  syncUrl: false,
})

const transcriptVisible = ref(false)
const transcriptId = ref<number>()
const transcriptTitle = ref('')

function openTranscript(row: SupportUserSession) {
  transcriptId.value = row.id
  transcriptTitle.value = `客服会话 #${row.id} · ${row.createdAt}`
  transcriptVisible.value = true
}

watch(
  () => props.userId,
  () => void handleSearch(),
)
</script>

<style scoped>
.user-support-tab__pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}
</style>
