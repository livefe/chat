<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import { login, sendLoginCode } from '@/api/user'
import { LOGIN_CODE_COUNTDOWN_SEC, LOGIN_COPY, MOBILE_PATTERN } from '@/constants/auth'
import { useChatStore } from '@/stores/chat'
import { useUserStore } from '@/stores/user'
import { safeRedirectPath } from '@/utils/safeRedirect'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

const mobile = ref('')
const code = ref('')
const agreed = ref(false)
const counting = ref(0)
const sending = ref(false)
const submitting = ref(false)

const mobileValid = computed(() => MOBILE_PATTERN.test(mobile.value))

let timer: number | undefined

function startCountdown() {
  if (timer) window.clearInterval(timer)
  counting.value = LOGIN_CODE_COUNTDOWN_SEC
  timer = window.setInterval(() => {
    counting.value -= 1
    if (counting.value <= 0 && timer) {
      window.clearInterval(timer)
      timer = undefined
    }
  }, 1000)
}

async function handleSendCode() {
  ;(document.activeElement as HTMLElement | null)?.blur()
  if (!mobileValid.value) {
    showToast(LOGIN_COPY.invalidMobile)
    return
  }
  if (counting.value > 0 || sending.value) return
  sending.value = true
  try {
    await sendLoginCode(mobile.value)
    showToast(LOGIN_COPY.codeSent)
    startCountdown()
  } finally {
    sending.value = false
  }
}

async function handleLogin() {
  ;(document.activeElement as HTMLElement | null)?.blur()
  if (!mobileValid.value) {
    showToast(LOGIN_COPY.invalidMobile)
    return
  }
  if (!code.value) {
    showToast(LOGIN_COPY.emptyCode)
    return
  }
  if (!agreed.value) {
    showToast(LOGIN_COPY.needAgreement)
    return
  }
  submitting.value = true
  try {
    const res = await login(mobile.value, code.value)

    chatStore.clearSessionState()

    userStore.setTokens('', '')
    userStore.setUser(null)
    userStore.setTokens(res.data.accessToken, res.data.refreshToken)
    userStore.setUser(res.data.user)
    const redirect = safeRedirectPath(router.currentRoute.value.query.redirect)
    router.replace(redirect)
  } finally {
    submitting.value = false
  }
}

function dismissKeyboardOutsideInput(e: Event) {
  const t = e.target as HTMLElement | null
  if (!t) return
  if (t.closest('input, textarea')) return
  if (t.closest('.van-field') && !t.closest('button, .code-btn, a')) return
  ;(document.activeElement as HTMLElement | null)?.blur()
}

onUnmounted(() => {
  if (timer) {
    window.clearInterval(timer)
    timer = undefined
  }
})
</script>

<template>
  <div class="login-page" @pointerdown.capture="dismissKeyboardOutsideInput">
    <div class="hero">
      <div class="title">聚点</div>
      <div class="subtitle">发现身边的人和事</div>
    </div>

    <form class="fields" autocomplete="off" @submit.prevent>
      <van-field
        v-model="mobile"
        label="手机号"
        placeholder="请输入手机号"
        type="digit"
        maxlength="11"
        name="login-mobile"
        autocomplete="off"
      />
      <van-field
        v-model="code"
        label="验证码"
        placeholder="请输入验证码"
        type="digit"
        maxlength="6"
        name="login-code"
        autocomplete="off"
      >
        <template #button>
          <button
            type="button"
            class="code-btn"
            :disabled="counting > 0 || sending"
            @click="handleSendCode"
          >
            {{
              counting > 0 ? `${counting}s` : sending ? LOGIN_COPY.sendingCode : LOGIN_COPY.getCode
            }}
          </button>
        </template>
      </van-field>
    </form>

    <van-button
      class="login-btn"
      type="primary"
      block
      :loading="submitting"
      :loading-text="LOGIN_COPY.loggingIn"
      @click="handleLogin"
    >
      {{ LOGIN_COPY.login }}
    </van-button>

    <label class="agreement">
      <span class="agreement-mark">
        <input v-model="agreed" class="agreement-check" type="checkbox" />
      </span>
      <span class="agreement-text">
        我已阅读并同意
        <span class="agreement-link" @click.prevent="router.push('/agreement/user')"
          >《用户协议》</span
        >
        与
        <span class="agreement-link" @click.prevent="router.push('/agreement/privacy')"
          >《隐私政策》</span
        >
      </span>
    </label>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  box-sizing: border-box;
  padding: 72px 16px 32px;
  background: var(--page-bg);
}

.hero {
  margin-bottom: 32px;
}

.hero .title {
  font-size: var(--font-hero);
  font-weight: 600;
  line-height: 1.25;
  color: var(--brand-color);
}

.hero .subtitle {
  margin-top: 8px;
  font-size: var(--font-secondary);
  line-height: 20px;
  color: var(--text-tertiary);
}

.fields {
  overflow: hidden;
  border-radius: var(--radius-md);
  background: var(--card-bg);
}

.fields :deep(.van-cell) {
  background: transparent;
  padding: 14px 16px;
  font-size: var(--font-body);
}

.fields :deep(.van-field__label) {
  width: 72px;
  margin-right: 0;
  color: var(--text-primary);
  font-size: var(--font-body);
}

.fields :deep(.van-field__control) {
  font-size: var(--font-body);
}

.fields :deep(.van-field__control::placeholder) {
  color: var(--text-tertiary);
}

.fields :deep(.van-cell::after) {
  left: 16px;
  right: 0;
  border-bottom-width: 1px;
  border-color: var(--border-color-soft);
  transform: none;
}

.code-btn {
  height: 30px;
  padding: 0 10px;
  border: 1px solid var(--brand-color);
  border-radius: 4px;
  background: transparent;
  color: var(--brand-color);
  font-size: var(--font-secondary);
  line-height: 28px;
  white-space: nowrap;
  cursor: pointer;
}

.code-btn:disabled {
  border-color: color-mix(in srgb, var(--brand-color) 35%, transparent);
  color: color-mix(in srgb, var(--brand-color) 35%, transparent);
  cursor: default;
}

.login-btn {
  margin-top: 24px;
  height: 44px;
  font-size: var(--font-body);
  font-weight: 600;
  border: none;
  box-shadow: none;
  border-radius: var(--radius-btn);
}

.login-btn.van-button--disabled {
  opacity: 1;
  background: color-mix(in srgb, var(--brand-color) 35%, transparent);
  color: #fff;
}

.agreement {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  margin-left: -30px;
  cursor: pointer;
  user-select: none;
}

.agreement-mark {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: flex-end;
  width: 44px;
  height: 44px;
}

.agreement-check {
  appearance: none;
  box-sizing: border-box;
  width: 14px;
  height: 14px;
  margin: 0;
  border: 1px solid var(--text-tertiary);
  border-radius: 4px;
  background: var(--card-bg);
  cursor: pointer;
}

.agreement-check:checked {
  border-color: var(--brand-color);
  background: var(--brand-color)
    url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 12 12'%3E%3Cpath fill='none' stroke='%23fff' stroke-width='2' stroke-linecap='round' stroke-linejoin='round' d='M2.5 6.2 4.8 8.5 9.5 3.5'/%3E%3C/svg%3E")
    center / 10px 10px no-repeat;
}

.agreement-check:focus {
  outline: none;
}

.agreement-text {
  font-size: var(--font-agree);
  line-height: 18px;
  color: var(--text-tertiary);
}

.agreement-link {
  color: var(--brand-color);
}
</style>
