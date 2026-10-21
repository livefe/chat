<template>
  <div class="admin-empty-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>无访问权限</span>
        </div>
      </template>

      <el-empty description="您没有权限访问此页面，请联系管理员分配菜单权限" />
      <div class="empty-actions">
        <el-button type="primary" @click="goHome">返回首页</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

import { HOME_PATH } from '@/router/dynamic'

const router = useRouter()

function goHome() {
  router.replace(HOME_PATH)
}
</script>
