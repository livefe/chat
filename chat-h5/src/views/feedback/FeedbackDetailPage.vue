<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { getMyFeedbackDetail, type FeedbackItem } from '@/api/feedback'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const route = useRoute()
const detail = ref<FeedbackItem | null>(null)
const loading = ref(true)

onMounted(async () => {
  const id = Number(route.params.id)
  if (!id) {
    showToast('参数错误')
    router.back()
    return
  }
  try {
    const res = await getMyFeedbackDetail(id)
    detail.value = res.data
  } catch (e: unknown) {
    showCatchError(e, '加载失败')
    router.back()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <SubPageLayout title="反馈详情">
    <van-loading v-if="loading" class="loading" />
    <template v-else-if="detail">
      <van-cell-group inset>
        <van-cell title="类型" :value="detail.typeText" />
        <van-cell title="状态" :value="detail.statusText" />
        <van-cell title="提交时间" :value="detail.createdAt" />
        <van-cell title="反馈内容" :label="detail.content" />
        <van-cell v-if="detail.contact" title="联系方式" :value="detail.contact" />
      </van-cell-group>

      <van-cell-group v-if="detail.images?.length" inset class="section">
        <van-cell title="截图" />
        <div class="images">
          <van-image
            v-for="(url, i) in detail.images"
            :key="i"
            width="80"
            height="80"
            fit="cover"
            :src="url"
          />
        </div>
      </van-cell-group>

      <van-cell-group v-if="detail.reply || detail.handledAt" inset class="section reply-block">
        <van-cell title="处理结果" />
        <van-cell v-if="detail.reply" title="处理说明" :label="detail.reply" />
        <van-cell v-if="detail.handledAt" title="处理时间" :value="detail.handledAt" />
      </van-cell-group>
      <van-cell-group v-else inset class="section">
        <van-cell title="处理结果" value="待处理，请耐心等待" />
      </van-cell-group>
    </template>
  </SubPageLayout>
</template>

<style scoped>
.loading {
  display: flex;
  justify-content: center;
  padding: 48px;
}
.section {
  margin-top: 12px;
}
.images {
  display: flex;
  gap: 8px;
  padding: 0 16px 12px;
  flex-wrap: wrap;
}
.reply-block :deep(.van-cell__label) {
  white-space: pre-wrap;
}
</style>
