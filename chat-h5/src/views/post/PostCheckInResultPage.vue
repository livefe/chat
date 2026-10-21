<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'
import {
  checkInStatusText,
  formatCheckInDisplayTime,
  parseCheckInResultPayload,
} from '@/utils/post/checkInResult'

const route = useRoute()
const router = useRouter()

const payload = computed(() => parseCheckInResultPayload(String(route.query.payload || '')))
const postId = computed(() => Number(payload.value?.postId || 0))

const isFail = computed(() => payload.value?.kind === 'fail')
const isAlready = computed(() => payload.value?.kind === 'already')
const isSuccess = computed(() => payload.value?.kind === 'success')
const statusText = computed(() => checkInStatusText(payload.value?.kind || 'success'))
const showUser = computed(() => !isFail.value && Boolean(payload.value?.userNickname))

const timeline = computed(() => {
  const p = payload.value
  if (!p || isFail.value) return []
  const items = []
  if (p.signupAt) {
    items.push({
      key: 'signup',
      label: '报名',
      time: formatCheckInDisplayTime(p.signupAt),
      meta: '',
      active: false,
    })
  }
  if (p.checkedInAtText) {
    items.push({
      key: 'checkin',
      label: '现场核验',
      time: formatCheckInDisplayTime(p.checkedInAtText),
      meta: p.operatorNickname ? `操作人：${p.operatorNickname}` : '',
      active: true,
    })
  }
  return items
})

function continueScan() {
  router.back()
}

function openSignups() {
  if (!postId.value) return
  router.replace(`/post/${postId.value}/signups`)
}
</script>

<template>
  <SubPageLayout full-height>
    <template #title>核验结果</template>
    <div class="result-page">
      <template v-if="payload">
        <div v-if="isFail" class="result-card result-card--fail">
          <div class="result-card__status">{{ statusText }}</div>
          <p class="result-card__msg">{{ payload.message || '请重新扫码，或在名单页手动核验' }}</p>
        </div>

        <template v-else>
          <div v-if="showUser" class="result-user">
            <img
              v-if="payload.userAvatarUrl"
              class="result-user__avatar"
              :src="payload.userAvatarUrl"
              alt=""
            />
            <div v-else class="result-user__avatar result-user__avatar--placeholder" />
            <div class="result-user__main">
              <div class="result-user__name">{{ payload.userNickname }}</div>
              <div v-if="payload.postTitle" class="result-user__post">{{ payload.postTitle }}</div>
            </div>
            <span
              class="result-user__tag"
              :class="{
                'result-user__tag--ok': isSuccess,
                'result-user__tag--already': isAlready,
              }"
            >
              {{ statusText }}
            </span>
          </div>

          <div v-if="timeline.length" class="timeline-card">
            <div class="timeline-card__title">参与记录</div>
            <div class="timeline">
              <div
                v-for="(item, index) in timeline"
                :key="item.key"
                class="timeline__item"
                :class="{ 'timeline__item--last': index === timeline.length - 1 }"
              >
                <div class="timeline__rail">
                  <span class="timeline__dot" :class="{ 'timeline__dot--active': item.active }" />
                  <span v-if="index < timeline.length - 1" class="timeline__line" />
                </div>
                <div class="timeline__body">
                  <div class="timeline__label">{{ item.label }}</div>
                  <div class="timeline__time">{{ item.time }}</div>
                  <div v-if="item.meta" class="timeline__meta">{{ item.meta }}</div>
                </div>
              </div>
            </div>
          </div>
        </template>
      </template>

      <div v-else class="result-card result-card--fail">
        <div class="result-card__status">数据缺失</div>
        <p class="result-card__msg">请返回重新扫码</p>
      </div>

      <div class="result-actions">
        <van-button type="primary" block round class="result-actions__main" @click="continueScan">
          继续扫下一个
        </van-button>
        <van-button
          v-if="postId"
          block
          round
          plain
          hairline
          class="result-actions__secondary"
          @click="openSignups"
        >
          查看报名名单
        </van-button>
      </div>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.result-page {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 12px 12px calc(16px + env(safe-area-inset-bottom));
  background: var(--page-bg);
}

.result-user,
.result-card,
.timeline-card {
  border-radius: 12px;
  background: var(--card-bg);
}

.result-user {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
}

.result-user__avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  background: #ececec;
}

.result-user__avatar--placeholder {
  background: linear-gradient(135deg, #e8eef8 0%, #dce7fb 100%);
}

.result-user__main {
  flex: 1;
  min-width: 0;
}

.result-user__name {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-user__post {
  margin-top: 4px;
  font-size: 13px;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-user__tag {
  flex-shrink: 0;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
}

.result-user__tag--ok {
  color: #047857;
  background: rgba(16, 185, 129, 0.12);
}

.result-user__tag--already {
  color: #b45309;
  background: rgba(180, 83, 9, 0.12);
}

.timeline-card {
  margin-top: 10px;
  padding: 16px;
}

.timeline-card__title {
  margin-bottom: 14px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
}

.timeline__item {
  display: flex;
  gap: 12px;
}

.timeline__item + .timeline__item {
  margin-top: 16px;
}

.timeline__rail {
  width: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}

.timeline__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #d9d9d9;
  flex-shrink: 0;
}

.timeline__dot--active {
  background: var(--brand-color);
  box-shadow: 0 0 0 4px rgba(76, 141, 255, 0.14);
}

.timeline__line {
  flex: 1;
  width: 2px;
  min-height: 24px;
  margin-top: 4px;
  background: #ececec;
}

.timeline__body {
  flex: 1;
  min-width: 0;
  padding-bottom: 2px;
}

.timeline__label {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
}

.timeline__time {
  margin-top: 4px;
  font-size: 13px;
  color: var(--text-secondary);
}

.timeline__meta {
  margin-top: 4px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.result-card {
  padding: 20px 16px;
}

.result-card--fail .result-card__status {
  color: #b91c1c;
}

.result-card__status {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}

.result-card__msg {
  margin: 10px 0 0;
  font-size: 14px;
  line-height: 1.55;
  color: var(--text-secondary);
  word-break: break-word;
}

.result-actions {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 24px;
}

.result-actions__main,
.result-actions__secondary {
  height: 44px;
  font-size: 16px;
}
</style>
