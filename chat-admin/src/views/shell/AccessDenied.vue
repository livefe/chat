<template>
  <div class="admin-empty-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>无访问权限</span>
        </div>
      </template>

      <el-empty :description="description" />
      <p v-if="permission" class="sub missing-perm">缺少权限编码：{{ permission }}</p>
      <div class="empty-actions">
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="switchAccount">切换账号</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { HOME_PATH } from '@/router/dynamic'

const route = useRoute()
const router = useRouter()

function queryString(key: string) {
  const v = route.query[key]
  return typeof v === 'string' ? v : ''
}

const title = computed(() => queryString('title'))
const permission = computed(() => queryString('permission'))
const description = computed(() =>
  title.value
    ? `当前账号无权访问「${title.value}」，请联系管理员分配该菜单权限`
    : '当前账号无权访问此页面，请联系管理员分配菜单权限',
)

function goHome() {
  void router.replace(HOME_PATH)
}

function switchAccount() {
  const from = queryString('from')
  void router.push({ path: '/login', query: { switch: '1', redirect: from || undefined } })
}
</script>

<style scoped>
.missing-perm {
  margin: -12px 0 12px;
  text-align: center;
}
</style>
