<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import QRCode from 'qrcode'
import { showFailToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { postCheckInCode } from '@/api/post'
import { showCatchError } from '@/utils/toast'

const route = useRoute()
const postId = computed(() => Number(route.params.id || 0))

const loading = ref(true)
const title = ref('')
const payload = ref('')
const checkedIn = ref(false)
const checkedInAtText = ref('')
const userNickname = ref('')
const signupAt = ref('')
const qrDataUrl = ref('')

const hintText = computed(() =>
  checkedIn.value ? '已完成核验，可凭此页向主办方确认' : '请出示给主办方扫一扫',
)

async function load() {
  if (!postId.value) return
  loading.value = true
  try {
    const res = await postCheckInCode(postId.value)
    const data = res.data
    title.value = data.postTitle
    payload.value = data.payload
    checkedIn.value = data.checkedIn
    checkedInAtText.value = data.checkedInAtText || ''
    userNickname.value = data.userNickname || ''
    signupAt.value = data.signupAt || ''
  } catch (e: unknown) {
    showCatchError(e, '加载失败', true)
  } finally {
    loading.value = false
  }
}

watch(
  payload,
  async (value) => {
    if (!value) {
      qrDataUrl.value = ''
      return
    }
    try {
      qrDataUrl.value = await QRCode.toDataURL(value, {
        width: 280,
        margin: 1,
        errorCorrectionLevel: 'M',
        color: { dark: '#111111', light: '#ffffff' },
      })
    } catch {
      qrDataUrl.value = ''
      showFailToast('二维码生成失败')
    }
  },
  { immediate: true },
)

onMounted(() => {
  void load()
})
</script>

<template>
  <SubPageLayout>
    <template #title>报名码</template>
    <div class="code-page">
      <van-skeleton title :row="6" :loading="loading">
        <div class="code-card">
          <div class="code-card__title">{{ title || '活动报名码' }}</div>
          <div class="code-card__hint">{{ hintText }}</div>
          <div class="code-card__qr" :class="{ 'code-card__qr--done': checkedIn }">
            <img v-if="qrDataUrl" :src="qrDataUrl" alt="报名码" />
          </div>
          <div class="code-card__status" :class="{ 'code-card__status--on': checkedIn }">
            {{ checkedIn ? '已核验' : '待核验' }}
          </div>
          <div v-if="checkedIn && checkedInAtText" class="code-card__time">
            核验时间 {{ checkedInAtText }}
          </div>
          <div v-if="userNickname || signupAt" class="code-card__footer">
            <div v-if="userNickname" class="code-card__row">
              <span class="code-card__label">报名昵称</span>
              <span class="code-card__value">{{ userNickname }}</span>
            </div>
            <div v-if="signupAt" class="code-card__row">
              <span class="code-card__label">报名时间</span>
              <span class="code-card__value">{{ signupAt }}</span>
            </div>
          </div>
        </div>
      </van-skeleton>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.code-page {
  min-height: 100%;
  padding: 24px 20px;
  background: var(--page-bg);
}

.code-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 28px 20px;
  border-radius: 16px;
  background: var(--card-bg);
}

.code-card__title {
  max-width: 100%;
  font-size: var(--font-page-title);
  font-weight: 600;
  color: var(--text-primary);
  text-align: center;
  word-break: break-word;
}

.code-card__hint {
  margin-top: 8px;
  font-size: var(--font-meta);
  color: var(--text-tertiary);
}

.code-card__qr {
  margin-top: 24px;
  width: 240px;
  height: 240px;
  padding: 12px;
  border-radius: 12px;
  background: #fff;
  box-sizing: border-box;
}

.code-card__qr img {
  width: 100%;
  height: 100%;
  display: block;
}

.code-card__qr--done {
  opacity: 0.55;
}

.code-card__status {
  margin-top: 20px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-secondary);
}

.code-card__status--on {
  color: #047857;
}

.code-card__time {
  margin-top: 6px;
  font-size: var(--font-meta);
  color: var(--text-tertiary);
}

.code-card__footer {
  width: 100%;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.code-card__row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.code-card__label {
  flex-shrink: 0;
  font-size: var(--font-meta);
  color: var(--text-tertiary);
}

.code-card__value {
  min-width: 0;
  font-size: var(--font-meta);
  color: var(--text-primary);
  text-align: right;
  word-break: break-word;
}
</style>
