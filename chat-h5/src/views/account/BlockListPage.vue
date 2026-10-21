<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { getBlockList, unblockUser, type BlockListItem } from '@/api/user'
import { openUserProfile } from '@/utils/user'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const list = ref<BlockListItem[]>([])
const loading = ref(false)
const finished = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const unblockingId = ref(0)

async function loadMore() {
  if (finished.value) return
  loading.value = true
  try {
    const res = await getBlockList({ page: page.value, pageSize })
    const items = res.data.list || []
    total.value = res.data.total
    list.value.push(...items)
    if (list.value.length >= total.value || items.length < pageSize) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch (e: unknown) {
    showCatchError(e, '加载失败', true)
    finished.value = true
  } finally {
    loading.value = false
  }
}

function openProfile(item: BlockListItem) {
  openUserProfile(router, item.userId, {
    from: 'settings',
    name: item.nickname,
    avatarUrl: item.avatarUrl,
  })
}

async function onUnblock(item: BlockListItem) {
  if (unblockingId.value) return
  try {
    await showConfirmDialog({
      title: '取消拉黑',
      message: `确定取消拉黑「${item.nickname}」吗？`,
    })
  } catch {
    return
  }
  unblockingId.value = item.userId
  try {
    await unblockUser(item.userId)
    list.value = list.value.filter((row) => row.userId !== item.userId)
    total.value = Math.max(0, total.value - 1)
    showToast('已取消拉黑')
  } catch (e: unknown) {
    showCatchError(e, '操作失败')
  } finally {
    unblockingId.value = 0
  }
}
</script>

<template>
  <SubPageLayout title="黑名单">
    <van-list
      v-model:loading="loading"
      :finished="finished"
      :finished-text="list.length ? '没有更多了' : ''"
      @load="loadMore"
    >
      <van-cell-group v-if="list.length" inset class="list-group">
        <van-cell v-for="item in list" :key="item.userId" center is-link @click="openProfile(item)">
          <template #icon>
            <UserAvatar :src="item.avatarUrl" :size="40" class="block-avatar" />
          </template>
          <template #title>
            <span class="block-name">{{ item.nickname }}</span>
          </template>
          <template #right-icon>
            <van-button
              size="small"
              plain
              hairline
              type="primary"
              :loading="unblockingId === item.userId"
              @click.stop="onUnblock(item)"
            >
              解除
            </van-button>
          </template>
        </van-cell>
      </van-cell-group>
      <van-empty v-else-if="finished && !list.length" description="暂无拉黑用户" />
    </van-list>
  </SubPageLayout>
</template>

<style scoped>
.list-group {
  margin-top: 12px;
}
.block-avatar {
  margin-right: 12px;
}
.block-name {
  font-size: 15px;
  color: var(--text-primary);
}
</style>
