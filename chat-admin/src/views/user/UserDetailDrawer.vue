<template>
  <el-drawer v-model="visible" title="用户详情" size="640px">
    <div v-loading="loading">
      <template v-if="detail">
        <div class="detail-actions">
          <el-button
            size="small"
            type="primary"
            @click="emit('messages', detail.id)"
            v-if="can(PERM.ChatMessage)"
          >
            查看全部消息
          </el-button>
        </div>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag v-if="detail.deactivated" type="info">已注销</el-tag>
            <el-tag v-else :type="detail.status === 1 ? 'success' : 'danger'">{{
              detail.statusText
            }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="手机号">{{ detail.mobile }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ detail.nickname || '-' }}</el-descriptions-item>
          <el-descriptions-item label="禁言">
            <el-tag v-if="detail.muteText" type="warning">{{ detail.muteText }}</el-tag>
            <span v-else>未禁言</span>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ detail.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ detail.updatedAt }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="detail.avatarUrl" class="detail-avatar">
          <el-avatar :src="detail.avatarUrl" :size="64" />
        </div>

        <el-divider content-position="left">活跃聊天区</el-divider>
        <el-table :data="detail.rooms" size="small" empty-text="暂无聊天区发言记录">
          <el-table-column prop="roomName" label="聊天区" min-width="140" />
          <el-table-column prop="messageCount" label="消息数" width="90" />
          <el-table-column prop="lastMessageAt" label="最后发言" width="170" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <div class="action-cell">
                <el-button
                  size="small"
                  link
                  type="primary"
                  @click="emit('messages', detail.id, row.roomId)"
                  v-if="can(PERM.ChatMessage)"
                >
                  消息
                </el-button>
                <el-button
                  size="small"
                  link
                  type="primary"
                  @click="emit('go-room', row.roomId)"
                  v-if="can(PERM.ChatRoom)"
                >
                  聊天区
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <el-divider content-position="left">最近登录</el-divider>
        <el-table :data="detail.loginRecords" size="small" empty-text="暂无登录记录">
          <el-table-column prop="createdAt" label="时间" width="170" />
          <el-table-column prop="ip" label="IP" width="130" />
          <el-table-column prop="statusText" label="结果" width="80" />
        </el-table>

        <el-divider content-position="left">注销记录</el-divider>
        <el-table :data="detail.deactivationLogs" size="small" empty-text="无注销记录">
          <el-table-column prop="createdAt" label="时间" width="170" />
          <el-table-column prop="sourceText" label="来源" width="100" />
          <el-table-column prop="reason" label="原因" show-overflow-tooltip />
        </el-table>
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import type { AdminUserDetail } from '@/api/user'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const visible = defineModel<boolean>({ required: true })

defineProps<{
  detail: AdminUserDetail | null
  loading: boolean
}>()

const emit = defineEmits<{
  messages: [userId: number, roomId?: number]
  'go-room': [roomId: number]
}>()

const { can } = usePermission()
</script>

<style scoped>
.detail-avatar {
  margin-top: 12px;
}
</style>
