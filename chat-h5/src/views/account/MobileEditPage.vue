<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import {
  changeMobile,
  sendChangeMobileNewCode,
  sendChangeMobileOldCode,
  verifyChangeMobileOldCode,
} from '@/api/user'
import { LOGIN_CODE_COUNTDOWN_SEC, MOBILE_PATTERN } from '@/constants/auth'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()

const step = ref<1 | 2>(1)
const currentMobile = ref(userStore.info?.mobile || '')
const ticket = ref('')
const oldCode = ref('')
const newMobile = ref('')
const newCode = ref('')
const submitting = ref(false)
const verifying = ref(false)

const oldCounting = ref(0)
const newCounting = ref(0)
const sendingOld = ref(false)
const sendingNew = ref(false)

let oldTimer: number | undefined
let newTimer: number | undefined

const pageTitle = computed(() => (step.value === 1 ? '验证手机号' : '绑定新手机号'))

const maskedMobile = computed(() => {
  const m = currentMobile.value
  if (m.length !== 11) return m || '—'
  return `${m.slice(0, 3)}****${m.slice(7)}`
})

const canGoNext = computed(
  () => oldCode.value.trim().length >= 4 && !sendingOld.value && !verifying.value,
)

const canSubmit = computed(() => {
  return (
    !!ticket.value &&
    MOBILE_PATTERN.test(newMobile.value) &&
    newCode.value.trim().length >= 4 &&
    !submitting.value
  )
})

onMounted(() => {
  currentMobile.value = userStore.info?.mobile || ''
})

onUnmounted(() => {
  if (oldTimer) window.clearInterval(oldTimer)
  if (newTimer) window.clearInterval(newTimer)
})

function handleBack() {
  if (step.value === 2) {
    step.value = 1
    return
  }
  router.back()
}

function startCountdown(kind: 'old' | 'new') {
  if (kind === 'old') {
    oldCounting.value = LOGIN_CODE_COUNTDOWN_SEC
    oldTimer = window.setInterval(() => {
      oldCounting.value -= 1
      if (oldCounting.value <= 0 && oldTimer) {
        window.clearInterval(oldTimer)
        oldTimer = undefined
      }
    }, 1000)
    return
  }
  newCounting.value = LOGIN_CODE_COUNTDOWN_SEC
  newTimer = window.setInterval(() => {
    newCounting.value -= 1
    if (newCounting.value <= 0 && newTimer) {
      window.clearInterval(newTimer)
      newTimer = undefined
    }
  }, 1000)
}

async function sendOldCode() {
  if (oldCounting.value > 0 || sendingOld.value) return
  sendingOld.value = true
  try {
    await sendChangeMobileOldCode()
    showToast('验证码已发送')
    startCountdown('old')
  } finally {
    sendingOld.value = false
  }
}

async function sendNewCode() {
  if (newCounting.value > 0 || sendingNew.value) return
  if (!ticket.value) {
    showToast('请先验证当前手机号')
    step.value = 1
    return
  }
  if (!MOBILE_PATTERN.test(newMobile.value)) {
    showToast('请输入正确的新手机号')
    return
  }
  if (newMobile.value === currentMobile.value) {
    showToast('新手机号不能与当前手机号相同')
    return
  }
  sendingNew.value = true
  try {
    await sendChangeMobileNewCode(newMobile.value, ticket.value)
    showToast('验证码已发送')
    startCountdown('new')
  } finally {
    sendingNew.value = false
  }
}

async function goNext() {
  if (!oldCode.value.trim()) {
    showToast('请输入验证码')
    return
  }
  verifying.value = true
  try {
    const res = await verifyChangeMobileOldCode(oldCode.value.trim())
    ticket.value = res.data?.ticket || ''
    if (!ticket.value) {
      showToast('验证失败，请重试')
      return
    }
    step.value = 2
  } finally {
    verifying.value = false
  }
}

async function submit() {
  if (!canSubmit.value) return
  if (!ticket.value) {
    showToast('请先验证当前手机号')
    step.value = 1
    return
  }
  if (!MOBILE_PATTERN.test(newMobile.value)) {
    showToast('请输入正确的新手机号')
    return
  }
  if (newMobile.value === currentMobile.value) {
    showToast('新手机号不能与当前手机号相同')
    return
  }
  submitting.value = true
  try {
    await changeMobile({
      ticket: ticket.value,
      newMobile: newMobile.value.trim(),
      newCode: newCode.value.trim(),
    })
    await userStore.fetchDetail()
    showToast('更换成功')
    router.back()
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <SubPageLayout :title="pageTitle" full-height class="mobile-edit" :on-back="handleBack">
    <div v-if="step === 1" class="mobile-edit__panel">
      <p class="mobile-edit__lead">为确认是你本人操作，请先验证当前手机号</p>
      <div class="mobile-edit__phone">{{ maskedMobile }}</div>

      <div class="mobile-edit__row">
        <input
          v-model="oldCode"
          class="mobile-edit__input"
          maxlength="6"
          type="tel"
          inputmode="numeric"
          placeholder="请输入验证码"
        />
        <button
          type="button"
          class="mobile-edit__code-btn"
          :disabled="oldCounting > 0 || sendingOld"
          @click="sendOldCode"
        >
          {{ oldCounting > 0 ? `${oldCounting}s` : sendingOld ? '发送中' : '获取验证码' }}
        </button>
      </div>

      <van-button
        class="page-btn mobile-edit__action"
        type="primary"
        block
        :disabled="!canGoNext"
        :loading="verifying"
        @click="goNext"
      >
        下一步
      </van-button>
    </div>

    <div v-else class="mobile-edit__panel">
      <p class="mobile-edit__lead">验证通过后，账号将绑定到新的手机号</p>

      <div class="mobile-edit__field">
        <input
          v-model="newMobile"
          class="mobile-edit__input mobile-edit__input--solo"
          maxlength="11"
          type="tel"
          inputmode="numeric"
          placeholder="请输入新手机号"
        />
      </div>

      <div class="mobile-edit__row">
        <input
          v-model="newCode"
          class="mobile-edit__input"
          maxlength="6"
          type="tel"
          inputmode="numeric"
          placeholder="请输入验证码"
        />
        <button
          type="button"
          class="mobile-edit__code-btn"
          :disabled="newCounting > 0 || sendingNew"
          @click="sendNewCode"
        >
          {{ newCounting > 0 ? `${newCounting}s` : sendingNew ? '发送中' : '获取验证码' }}
        </button>
      </div>

      <van-button
        class="page-btn mobile-edit__action"
        type="primary"
        block
        :disabled="!canSubmit"
        :loading="submitting"
        @click="submit"
      >
        确认更换
      </van-button>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.mobile-edit__panel {
  padding: 28px 20px 32px;
}

.mobile-edit__lead {
  margin: 0;
  color: var(--text-tertiary);
  font-size: 14px;
  line-height: 1.5;
}

.mobile-edit__phone {
  margin-top: 20px;
  margin-bottom: 28px;
  color: var(--text-primary);
  font-size: 28px;
  font-weight: 600;
  letter-spacing: 0.04em;
  line-height: 1.2;
}

.mobile-edit__field {
  margin-top: 8px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--border-color-soft);
}

.mobile-edit__row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
  border-bottom: 1px solid var(--border-color-soft);
}

.mobile-edit__input {
  flex: 1;
  min-width: 0;
  height: 48px;
  padding: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 17px;
}

.mobile-edit__input--solo {
  width: 100%;
}

.mobile-edit__input::placeholder {
  color: var(--text-tertiary);
}

.mobile-edit__code-btn {
  flex-shrink: 0;
  height: 32px;
  padding: 0 4px;
  border: none;
  background: transparent;
  color: var(--brand-color);
  font-size: 15px;
  white-space: nowrap;
}

.mobile-edit__code-btn:disabled {
  color: var(--text-tertiary);
}

.mobile-edit__action {
  margin-top: 40px;
}
</style>
