<template>
  <div class="admin-empty-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>页面不存在</span>
        </div>
      </template>

      <el-empty description="页面不存在，或当前账号没有该菜单权限" />
      <div class="empty-actions">
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="goBack">返回上一页</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

import { HOME_PATH } from '@/router/dynamic'

const router = useRouter()

function goHome() {
  void router.push(HOME_PATH)
}

function goBack() {
  if (window.history.length > 1) router.back()
  else goHome()
}
</script>
