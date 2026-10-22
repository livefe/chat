<template>
  <div class="login-container">
    <div class="login-form">
      <div class="login-header">
        <h2>管理后台</h2>
        <p class="login-sub">请使用员工手机号登录</p>
      </div>

      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        class="login-form-content"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="mobile">
          <el-input
            v-model="form.mobile"
            placeholder="请输入手机号"
            size="large"
            prefix-icon="Iphone"
            maxlength="11"
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item prop="code">
          <div class="code-input">
            <el-input
              v-model="form.code"
              placeholder="请输入验证码"
              size="large"
              prefix-icon="Message"
              maxlength="6"
              @keyup.enter="handleLogin"
            />
            <el-button
              size="large"
              :loading="sendingCode"
              :disabled="countdown > 0 || sendingCode"
              @click="handleSendCode"
              class="send-btn"
            >
              {{ countdown > 0 ? `${countdown}s后重试` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            native-type="submit"
            class="login-btn"
            :loading="loading"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { showApiError } from '@/utils/apiError'
import { useEmployeeStore } from '@/stores/employee'
import { MOBILE_PATTERN, MOBILE_RULE } from '@/constants'
import { HOME_PATH } from '@/router/dynamic'
import { safeRedirect } from '@/utils/navigation'
import { sendLoginCode } from '@/api/sms'

const route = useRoute()
const router = useRouter()
const employeeStore = useEmployeeStore()

const formRef = ref()
const loading = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const form = reactive({
  mobile: '',
  code: '',
})

const rules = {
  mobile: [{ required: true, message: '请输入手机号', trigger: 'blur' }, MOBILE_RULE],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码为6位数字', trigger: 'blur' },
  ],
}

function clearCountdownTimer() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function startCountdown() {
  clearCountdownTimer()
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearCountdownTimer()
    }
  }, 1000)
}

async function handleSendCode() {
  if (sendingCode.value || countdown.value > 0) return
  if (!MOBILE_PATTERN.test(form.mobile)) {
    ElMessage.warning('请输入正确的手机号')
    return
  }
  sendingCode.value = true
  try {
    await sendLoginCode(form.mobile)
    ElMessage.success('验证码发送成功')
    startCountdown()
  } catch (e) {
    showApiError(e, '验证码发送失败')
  } finally {
    sendingCode.value = false
  }
}

async function handleLogin() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    await employeeStore.login(form.mobile, form.code)
    ElMessage.success('登录成功')
    await router.replace(safeRedirect(route.query.redirect) ?? HOME_PATH)
  } catch (e) {
    showApiError(e, '登录失败')
  } finally {
    loading.value = false
  }
}

onBeforeUnmount(clearCountdownTimer)
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(180deg, #f0f0f0 0%, #f5f5f5 55%, #ffffff 100%);
}

.login-form {
  width: 400px;
  padding: 36px 40px 40px;
  background: var(--el-bg-color);
  border-radius: 16px;
  border: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 4px 24px rgba(76, 141, 255, 0.12);
}

.login-header {
  text-align: center;
  margin-bottom: 28px;
}

.login-header h2 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 22px;
  font-weight: 600;
}

.login-sub {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.login-form-content {
  width: 100%;
}

.code-input {
  display: flex;
  gap: 10px;
  width: 100%;
}

.code-input :deep(.el-input) {
  flex: 1;
}

.send-btn {
  width: 130px;
  flex-shrink: 0;
  padding: 0 12px;
}

.login-btn {
  width: 100%;
}
</style>
