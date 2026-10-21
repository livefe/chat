<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog, showToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { deactivateAccount, sendDeactivationCode } from '@/api/user'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const code = ref('')
const agreed = ref(false)
const counting = ref(0)
const sending = ref(false)
const submitting = ref(false)

const mobile = computed(() => userStore.info?.mobile ?? '')

let timer: number | undefined

onUnmounted(() => {
  if (timer) {
    window.clearInterval(timer)
    timer = undefined
  }
})

async function handleSendCode() {
  if (counting.value > 0 || sending.value) return
  sending.value = true
  try {
    await sendDeactivationCode()
    showToast('验证码已发送')
    counting.value = 60
    timer = window.setInterval(() => {
      counting.value -= 1
      if (counting.value <= 0 && timer) {
        window.clearInterval(timer)
        timer = undefined
      }
    }, 1000)
  } finally {
    sending.value = false
  }
}

async function submitDeactivate() {
  if (!agreed.value) {
    showToast('请先阅读并同意注销须知')
    return
  }
  if (!code.value.trim()) {
    showToast('请输入验证码')
    return
  }

  showDialog({
    title: '确认注销',
    message: '注销后账号数据将被删除且无法恢复，确定继续吗？',
    showCancelButton: true,
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    confirmButtonColor: 'var(--text-danger)',
    messageAlign: 'left',
  })
    .then(async () => {
      submitting.value = true
      try {
        await deactivateAccount(code.value.trim())
        showToast('账号已注销')
        userStore.logout()
        router.replace('/login')
      } finally {
        submitting.value = false
      }
    })
    .catch(() => {})
}
</script>

<template>
  <SubPageLayout title="注销账号">
    <div class="deactivate-page">
      <div class="notice">
        <p class="notice-title">注销须知</p>
        <p>注销后个人信息将被删除且无法恢复，请谨慎操作</p>
      </div>

      <van-cell-group inset>
        <van-field label="手机号" :model-value="mobile" readonly />
        <van-field
          v-model="code"
          label="验证码"
          placeholder="请输入验证码"
          maxlength="6"
          type="digit"
        >
          <template #button>
            <van-button
              size="small"
              type="primary"
              plain
              :disabled="counting > 0"
              :loading="sending"
              @click="handleSendCode"
            >
              {{ counting > 0 ? `${counting}s` : '获取验证码' }}
            </van-button>
          </template>
        </van-field>
      </van-cell-group>

      <van-checkbox v-model="agreed" icon-size="14px" class="agree-check">
        我已阅读并同意上述注销须知
      </van-checkbox>

      <van-button
        class="page-btn"
        type="danger"
        block
        :disabled="!agreed"
        :loading="submitting"
        @click="submitDeactivate"
      >
        确认注销账号
      </van-button>
    </div>
  </SubPageLayout>
</template>

<style scoped>
.deactivate-page {
  padding: 16px 0 32px;
}
.notice {
  margin: 0 16px 16px;
  padding: 14px 16px;
  background: var(--card-bg);
  border-radius: 12px;
  font-size: 14px;
  color: #666666;
  line-height: 1.6;
}
.notice-title {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 600;
  color: #323232;
}
.notice p:last-child {
  margin: 0;
}
.agree-check {
  margin: 16px;
  font-size: 12px;
  color: #666666;
}
</style>
