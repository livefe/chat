<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { closeToast, showFailToast, showLoadingToast, showSuccessToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import type { PostSignupItem } from '@/api/post'
import { usePostSignupList } from '@/composables/usePostSignupList'
import { showCatchError } from '@/utils/toast'
import PostSignupCard from './components/PostSignupCard.vue'
import PostSignupFilters from './components/PostSignupFilters.vue'
import PostSignupRemarkSheet from './components/PostSignupRemarkSheet.vue'

const route = useRoute()
const postId = computed(() => Number(route.params.id || 0))

const {
  list,
  loading,
  finished,
  refreshing,
  statusFilter,
  checkInFilter,
  needCheckIn,
  counts,
  loadPostTitle,
  loadMore,
  reloadList,
  refresh,
  setFilters,
  canManualCheckIn,
  manualCheckIn,
  saveRemark,
  fetchAllSignups,
  downloadSignups,
} = usePostSignupList(postId, {
  onLoadError: (e) => showCatchError(e, '加载失败', true),
})

const exporting = ref(false)
const checkingId = ref(0)
const remarkTarget = ref<PostSignupItem | null>(null)

async function onExport() {
  if (!postId.value || exporting.value) return
  exporting.value = true
  showLoadingToast({ message: '正在导出…', forbidClick: true, duration: 0 })
  try {
    const rows = await fetchAllSignups()
    closeToast()
    if (!rows.length) {
      showFailToast('暂无数据可导出')
      return
    }
    downloadSignups(rows)
    showSuccessToast('已导出')
  } catch (e: unknown) {
    closeToast()
    showCatchError(e, '导出失败', true)
  } finally {
    exporting.value = false
  }
}

async function onManualCheckIn(item: PostSignupItem) {
  if (!postId.value || !canManualCheckIn(item) || checkingId.value) return
  checkingId.value = item.id
  try {
    const already = await manualCheckIn(item)
    showSuccessToast(already ? '已核验' : '核验成功')
  } catch (e: unknown) {
    showCatchError(e, '核验失败', true)
  } finally {
    checkingId.value = 0
  }
}

async function onSaveRemark(item: PostSignupItem, remark: string) {
  if (!postId.value) return false
  try {
    await saveRemark(item, remark)
    showSuccessToast('已保存')
    return true
  } catch (e: unknown) {
    showCatchError(e, '保存失败', true)
    return false
  }
}

onMounted(() => {
  void loadPostTitle()
  void reloadList()
})
</script>

<template>
  <SubPageLayout full-height>
    <template #title>报名名单</template>
    <template #right>
      <button type="button" class="export-btn" :disabled="exporting" @click="onExport">
        {{ exporting ? '导出中' : '导出' }}
      </button>
    </template>

    <PostSignupFilters
      :counts="counts"
      :status="statusFilter"
      :check-in="checkInFilter"
      :need-check-in="needCheckIn"
      @change="setFilters"
    />

    <van-pull-refresh v-model="refreshing" class="list-scroll" @refresh="refresh">
      <van-list
        :loading="loading && !refreshing"
        :finished="finished"
        :finished-text="list.length ? '没有更多了' : ''"
        :immediate-check="false"
        @update:loading="
          (v) => {
            if (!refreshing) loading = v
          }
        "
        @load="loadMore"
      >
        <div v-if="list.length" class="signup-list">
          <PostSignupCard
            v-for="item in list"
            :key="item.id"
            :item="item"
            :need-check-in="needCheckIn"
            :can-check-in="canManualCheckIn(item)"
            :checking="checkingId === item.id"
            @edit-remark="remarkTarget = item"
            @check-in="onManualCheckIn(item)"
          />
        </div>
        <van-empty v-else-if="finished && !list.length" description="暂无报名记录" />
      </van-list>
    </van-pull-refresh>

    <PostSignupRemarkSheet
      :target="remarkTarget"
      :save="onSaveRemark"
      @close="remarkTarget = null"
    />
  </SubPageLayout>
</template>

<style scoped>
.export-btn {
  border: none;
  background: transparent;
  padding: 0 4px;
  font-size: 14px;
  line-height: 1;
  color: var(--text-primary);
}

.export-btn:disabled {
  color: var(--text-tertiary);
}

.list-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.signup-list {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
