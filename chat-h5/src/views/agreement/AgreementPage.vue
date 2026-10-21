<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'
import { isInAppShell } from '@/utils/appShell'

const props = defineProps<{ doc: 'user' | 'privacy' }>()
const route = useRoute()
const inApp = computed(() => isInAppShell(route.query))

const docs = {
  user: {
    title: '用户协议',
    sections: [
      {
        heading: '一、协议说明',
        body: '欢迎使用本产品与相关服务。本协议是您与平台之间关于使用服务所订立的协议。您在注册、登录、使用服务前，应仔细阅读并充分理解本协议全部内容。',
      },
      {
        heading: '二、账号与使用',
        body: '您应当保证注册信息真实、准确、完整，并妥善保管账号信息。因您个人原因导致账号泄露或损失的，需自行承担相应责任。',
      },
      {
        heading: '三、用户行为规范',
        body: '您不得利用本服务发布违法违规、侵害他人权益或违反公序良俗的内容。平台有权根据法律法规和平台规则对违规内容及账号采取处置措施。',
      },
      {
        heading: '四、免责声明',
        body: '在法律允许范围内，平台将尽力保障服务稳定运行，但不对因不可抗力、网络故障或第三方原因导致的服务中断承担责任。',
      },
      {
        heading: '五、协议变更',
        body: '平台有权根据业务需要对本协议进行更新。更新后会通过合理方式提示，您继续使用服务即视为接受更新后的协议内容。',
      },
    ],
  },
  privacy: {
    title: '隐私政策',
    sections: [
      {
        heading: '一、信息收集',
        body: '为了向您提供基础服务，我们可能收集必要信息，如手机号、设备信息、操作日志等。我们仅在实现服务功能所需范围内收集信息。',
      },
      {
        heading: '二、信息使用',
        body: '我们会将收集的信息用于身份验证、账号安全、功能优化、客户支持及法律法规要求的其他用途，不会超出合法、正当、必要范围使用。',
      },
      {
        heading: '三、信息共享',
        body: '未经您同意，我们不会向无关第三方提供您的个人信息。但在法律法规要求、维护公共利益或保护平台及用户权益所必需时除外。',
      },
      {
        heading: '四、信息存储与安全',
        body: '我们采取合理可行的安全措施保护您的信息，防止未经授权访问、披露、篡改或丢失。您也应妥善保管账号与验证码信息。',
      },
      {
        heading: '五、您的权利',
        body: '您有权查询、更正或删除您的个人信息，并可通过平台提供的方式申请注销账号。法律法规另有规定的，从其规定。',
      },
    ],
  },
} as const

const content = docs[props.doc]

onMounted(() => {
  document.title = content.title
})
</script>

<template>
  <SubPageLayout v-if="!inApp" :title="content.title">
    <div class="agreement-content">
      <template v-for="(section, i) in content.sections" :key="i">
        <h3>{{ section.heading }}</h3>
        <p>{{ section.body }}</p>
      </template>
    </div>
  </SubPageLayout>
  <div v-else class="agreement-app">
    <template v-for="(section, i) in content.sections" :key="i">
      <h3>{{ section.heading }}</h3>
      <p>{{ section.body }}</p>
    </template>
  </div>
</template>

<style scoped>
.agreement-app {
  min-height: 100vh;
  padding: 16px;
  background: #fff;
  color: #323232;
  line-height: 1.75;
  font-size: 14px;
  box-sizing: border-box;
}
.agreement-content {
  margin: 12px;
  padding: 16px;
  border-radius: 12px;
  background: var(--card-bg);
  color: #323232;
  line-height: 1.75;
  font-size: 14px;
}
.agreement-content h3,
.agreement-app h3 {
  margin: 0 0 8px;
  font-size: 15px;
  color: #1f1f1f;
}
.agreement-content p + h3,
.agreement-app p + h3 {
  margin-top: 14px;
}
.agreement-content p,
.agreement-app p {
  margin: 0;
}
</style>
