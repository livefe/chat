<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import SubPageLayout from '@/components/SubPageLayout.vue'
import { getMyFeedbackList, type FeedbackItem } from '@/api/feedback'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const list = ref<FeedbackItem[]>([])
const loading = ref(false)
const finished = ref(false)
const page = ref(1)
const pageSize = 10
const total = ref(0)

function statusClass(status: number) {
  if (status === 1) return 'resolved'
  if (status === 2) return 'closed'
  return 'pending'
}

async function loadMore() {
  if (finished.value) return
  loading.value = true
  try {
    const res = await getMyFeedbackList({ page: page.value, pageSize })
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
</script>

<template>
  <SubPageLayout title="我的反馈">
    <van-list
      v-model:loading="loading"
      :finished="finished"
      finished-text="没有更多了"
      @load="loadMore"
    >
      <van-cell-group v-if="list.length" inset class="list-group">
        <van-cell
          v-for="item in list"
          :key="item.id"
          is-link
          :title="item.typeText"
          :label="item.content"
          @click="router.push(`/me/feedback/records/${item.id}`)"
        >
          <template #value>
            <span :class="['status', statusClass(item.status)]">{{ item.statusText }}</span>
          </template>
        </van-cell>
      </van-cell-group>
      <van-empty v-else-if="finished && !list.length" description="暂无反馈记录" />
    </van-list>
  </SubPageLayout>
</template>

<style scoped>
.list-group {
  margin-top: 12px;
}
.status {
  font-size: 12px;
}
.status.pending {
  color: var(--text-danger);
}
.status.resolved {
  color: var(--text-reward);
}
.status.closed {
  color: var(--text-muted);
}
</style>
