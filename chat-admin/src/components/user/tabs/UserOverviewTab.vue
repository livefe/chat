<template>
  <div class="user-tab">
    <el-descriptions :column="2" border size="small">
      <el-descriptions-item label="昵称">{{ detail.nickname || '-' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag v-if="detail.deactivated" type="info" size="small">已注销</el-tag>
        <el-tag v-else :type="detail.status === 1 ? 'success' : 'danger'" size="small">
          {{ detail.statusText }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="禁言">
        <el-tag v-if="detail.muteText" type="warning" size="small">{{ detail.muteText }}</el-tag>
        <span v-else>未禁言</span>
      </el-descriptions-item>
      <el-descriptions-item label="最近登录">{{ detail.lastLoginAt || '-' }}</el-descriptions-item>
      <el-descriptions-item label="注册时间">{{ detail.createdAt }}</el-descriptions-item>
      <el-descriptions-item label="资料更新">{{ detail.updatedAt }}</el-descriptions-item>
    </el-descriptions>

    <div class="user-tab__section">活跃聊天区</div>
    <el-table :data="detail.rooms" size="small" empty-text="暂无聊天区发言记录">
      <el-table-column prop="roomName" label="聊天区" min-width="140" />
      <el-table-column prop="messageCount" label="消息数" width="90" />
      <el-table-column prop="lastMessageAt" label="最后发言" width="170" />
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <div class="action-cell">
            <el-button
              v-if="can(PERM.ChatMessage)"
              size="small"
              link
              type="primary"
              @click="
                go(PERM.ChatMessage, {
                  userId: String(userId),
                  roomId: String(row.roomId),
                  scope: String(MessageScope.Room),
                })
              "
            >
              消息
            </el-button>
            <el-button
              v-if="can(PERM.ChatRoom)"
              size="small"
              link
              type="primary"
              @click="go(PERM.ChatRoom, { id: String(row.roomId) })"
            >
              聊天区
            </el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <div class="user-tab__section">登录记录</div>
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty :error="error" text="暂无登录记录" @retry="fetchList" />
      </template>
      <el-table-column prop="createdAt" label="时间" width="160" />
      <el-table-column label="结果" width="70">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.statusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="loginType" label="方式" width="90" />
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column prop="userAgent" label="设备" min-width="160" show-overflow-tooltip />
      <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
    </el-table>
    <el-pagination
      v-if="total > query.pageSize"
      v-model:current-page="query.page"
      class="user-tab__pager"
      small
      layout="total, prev, pager, next"
      :page-size="query.pageSize"
      :total="total"
      @current-change="fetchList"
    />

    <template v-if="detail.deactivationLogs.length">
      <div class="user-tab__section">注销记录</div>
      <el-table :data="detail.deactivationLogs" size="small">
        <el-table-column prop="createdAt" label="时间" width="160" />
        <el-table-column prop="sourceText" label="来源" width="100" />
        <el-table-column prop="mobile" label="原手机号" width="130" />
        <el-table-column prop="reason" label="原因" show-overflow-tooltip />
      </el-table>
    </template>
  </div>
</template>

<script setup lang="ts">
import { getUserLoginRecords, type AdminUserDetail } from '@/api/user'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { useDrawerNav } from './useDrawerNav'

const props = defineProps<{ userId: number; detail: AdminUserDetail }>()

const { can } = usePermission()
const { go } = useDrawerNav()

const { query, list, total, loading, error, fetchList } = useListQuery({
  fetch: (q) => getUserLoginRecords(q),
  defaults: { page: 1, pageSize: 10, userId: props.userId },
  errorMessage: '获取登录记录失败',
  syncUrl: false,
})
</script>
