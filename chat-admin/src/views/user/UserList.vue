<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <template v-if="selected.length">
              <span class="sub">已选 {{ selected.length }} 人</span>
              <el-button
                v-if="can(PERM.UserAccountMute)"
                type="warning"
                plain
                @click="actions.openMute(selected)"
              >
                批量禁言
              </el-button>
              <el-button
                v-if="can(PERM.UserAccountStatus)"
                type="danger"
                plain
                @click="actions.batchDisable(selected)"
              >
                批量禁用
              </el-button>
            </template>
            <el-button
              v-if="can(PERM.UserAccountCreate)"
              type="primary"
              @click="actions.openCreate()"
            >
              新增用户
            </el-button>
          </div>
        </div>
      </template>

      <UserFilter
        v-model:id="query.id"
        v-model:mobile="query.mobile"
        v-model:nickname="query.nickname"
        v-model:status="query.status"
        v-model:muted="query.muted"
        v-model:start-time="query.startTime"
        v-model:end-time="query.endTime"
        @search="handleSearch"
        @reset="handleReset"
      />

      <UserTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @selection-change="(rows) => (selected = rows)"
        @page-change="fetchList"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { getUserList, type UserItem } from '@/api/user'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { useUserActions } from '@/composables/useUserActions'
import { PERM } from '@/constants/permission'
import { useUserDrawerStore } from '@/stores/userDrawer'
import UserFilter from './UserFilter.vue'
import UserTable from './UserTable.vue'

const route = useRoute()
const drawer = useUserDrawerStore()
const actions = useUserActions()
const { can } = usePermission()
const selected = ref<UserItem[]>([])

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getUserList(q),
  defaults: {
    page: 1,
    pageSize: 20,
    id: undefined as number | undefined,
    mobile: '',
    nickname: '',
    status: undefined as number | undefined,
    muted: undefined as boolean | undefined,
    startTime: '',
    endTime: '',
  },
  schema: {
    id: 'number',
    mobile: 'string',
    nickname: 'string',
    status: 'number',
    muted: 'boolean',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取用户列表失败',
})

watch(() => drawer.changed, fetchList)
</script>
