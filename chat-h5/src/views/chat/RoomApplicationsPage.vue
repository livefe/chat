<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { getMyRoomApplications, type RoomApplicationItem } from '@/api/room'
import { roomPlaceTypeLabel } from '@/utils/chat/chatRoomPlaceType'
import { showCatchError } from '@/utils/toast'

const STATUS_PENDING = 0
const STATUS_ENABLED = 1
const STATUS_REJECTED = 2

const router = useRouter()
const loading = ref(false)
const refreshing = ref(false)
const list = ref<RoomApplicationItem[]>([])

function statusClass(status: number) {
  if (status === STATUS_PENDING) return 'is-pending'
  if (status === STATUS_ENABLED) return 'is-enabled'
  if (status === STATUS_REJECTED) return 'is-rejected'
  return 'is-disabled'
}

function metaText(item: RoomApplicationItem) {
  const type = item.placeTypeText?.trim() || roomPlaceTypeLabel(item.placeType)
  return [type, item.createdAt].filter(Boolean).join(' · ')
}

async function loadList(opts?: { silent?: boolean }) {
  if (!opts?.silent) loading.value = true
  try {
    const res = await getMyRoomApplications()
    list.value = res.data.list || []
  } catch (e: unknown) {
    list.value = []
    showCatchError(e, '加载失败', true)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

async function onRefresh() {
  refreshing.value = true
  await loadList({ silent: true })
}

function onItemClick(item: RoomApplicationItem) {
  if (item.status === STATUS_ENABLED) {
    router.push(`/room/${item.id}`)
    return
  }
  showToast(item.statusText || '审核中')
}

onMounted(() => {
  void loadList()
})
</script>

<template>
  <SubPageLayout title="申请记录" full-height>
    <van-pull-refresh v-model="refreshing" class="apps-page" @refresh="onRefresh">
      <div v-if="loading && list.length === 0" class="tab-page__status">
        <van-loading size="24px" vertical>加载中…</van-loading>
      </div>
      <van-empty v-else-if="list.length === 0" description="暂无记录" />
      <div v-else class="apps-list">
        <button
          v-for="item in list"
          :key="item.id"
          type="button"
          class="apps-item"
          @click="onItemClick(item)"
        >
          <span class="apps-item__body">
            <span class="apps-item__name">{{ item.name || '聊天区' }}</span>
            <span class="apps-item__meta">{{ metaText(item) }}</span>
          </span>
          <span class="apps-item__status" :class="statusClass(item.status)">
            {{ item.statusText || '未知' }}
          </span>
        </button>
      </div>
    </van-pull-refresh>
  </SubPageLayout>
</template>

<style scoped>
.apps-page {
  min-height: 100%;
  box-sizing: border-box;
}

.apps-list {
  padding: 0 16px;
}

.apps-item {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 14px 0;
  border: none;
  border-bottom: 1px solid var(--border-color);
  background: transparent;
  text-align: left;
  font: inherit;
  color: inherit;
  cursor: pointer;
}

.apps-item:last-child {
  border-bottom: none;
}

.apps-item__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.apps-item__name {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.apps-item__meta {
  font-size: 12px;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.apps-item__status {
  flex-shrink: 0;
  margin-left: 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.apps-item__status.is-pending {
  color: #c9952a;
}

.apps-item__status.is-enabled {
  color: #2f9b63;
}

.apps-item__status.is-rejected {
  color: #d14343;
}
</style>
