<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="selected.length" class="card-header-actions">
            <span class="sub">已选 {{ selected.length }} 条</span>
            <el-button type="danger" plain @click="deleteRef?.open(selected)">批量删除</el-button>
          </div>
        </div>
      </template>

      <MessageFilter
        v-model:id="query.id"
        v-model:user-id="query.userId"
        v-model:peer-user-id="query.peerUserId"
        v-model:room-id="query.roomId"
        v-model:post-id="query.postId"
        v-model:scope="query.scope"
        v-model:keyword="query.keyword"
        v-model:start-time="query.startTime"
        v-model:end-time="query.endTime"
        @search="handleSearch"
        @reset="handleReset"
      />

      <MessageTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @context="(row) => contextRef?.open(row.id)"
        @delete="(row) => deleteRef?.open([row])"
        @selection-change="(rows) => (selected = rows)"
        @page-change="fetchList"
      />
    </el-card>

    <MessageContextDialog ref="contextRef" />
    <MessageDeleteDialog ref="deleteRef" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getMessageList, type AdminMessageItem } from '@/api/message'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import MessageContextDialog from './MessageContextDialog.vue'
import MessageDeleteDialog from './MessageDeleteDialog.vue'
import MessageFilter from './MessageFilter.vue'
import MessageTable from './MessageTable.vue'

const route = useRoute()
const router = useRouter()
const { can } = usePermission()
const contextRef = ref<InstanceType<typeof MessageContextDialog>>()
const deleteRef = ref<InstanceType<typeof MessageDeleteDialog>>()
const selected = ref<AdminMessageItem[]>([])

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getMessageList(q),
  defaults: {
    page: 1,
    pageSize: 20,
    id: undefined as number | undefined,
    userId: undefined as number | undefined,
    peerUserId: undefined as number | undefined,
    roomId: undefined as number | undefined,
    postId: undefined as number | undefined,
    scope: undefined as number | undefined,
    keyword: '',
    startTime: '',
    endTime: '',
  },
  schema: {
    id: 'number',
    userId: 'number',
    peerUserId: 'number',
    roomId: 'number',
    postId: 'number',
    scope: 'number',
    keyword: 'string',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取消息列表失败',
})

function openContextFromRoute() {
  const raw = route.query.contextId
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isInteger(id) || id <= 0) return
  const rest = { ...route.query }
  delete rest.contextId
  void router.replace({ path: route.path, query: rest })
  if (can(PERM.ChatMessageContext)) void contextRef.value?.open(id)
}

onMounted(openContextFromRoute)
watch(() => route.query.contextId, openContextFromRoute)
</script>
