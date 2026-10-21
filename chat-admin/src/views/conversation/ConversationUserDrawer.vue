<template>
  <el-drawer v-model="visible" title="用户详情" size="520px">
    <div v-loading="loading">
      <template v-if="detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ detail.mobile }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ detail.nickname || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="detail.status === 1 ? 'success' : 'danger'">{{
              detail.statusText
            }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="禁言">
            <el-tag v-if="detail.muteText" type="warning">{{ detail.muteText }}</el-tag>
            <span v-else>未禁言</span>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ detail.createdAt }}</el-descriptions-item>
        </el-descriptions>
        <div class="drawer-actions">
          <el-button
            v-if="can(PERM.ChatMessage)"
            size="small"
            type="primary"
            @click="emit('messages', detail.id)"
          >
            查看全部消息
          </el-button>
        </div>
      </template>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import type { AdminUserDetail } from '@/api/user'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'

defineProps<{
  detail: AdminUserDetail | null
  loading: boolean
}>()

const visible = defineModel<boolean>({ required: true })
const emit = defineEmits<{ messages: [userId: number] }>()

const { can } = usePermission()
</script>

<style scoped>
.drawer-actions {
  margin-top: 16px;
}
</style>
