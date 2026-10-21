<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import TabBarIcons, { type TabBarIconName } from './components/TabBarIcons.vue'
import { useChatStore } from '@/stores/chat'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const chatStore = useChatStore()
const userStore = useUserStore()

const tabs: { routeName: string; icon: TabBarIconName; label: string }[] = [
  { routeName: 'Chat', icon: 'chat', label: '附近' },
  { routeName: 'PostFeed', icon: 'postFeed', label: '发现' },
  { routeName: 'ConversationList', icon: 'conversation', label: '消息' },
  { routeName: 'Me', icon: 'me', label: '我的' },
]

const active = computed<number>(() =>
  Math.max(
    0,
    tabs.findIndex((tab) => tab.routeName === route.name),
  ),
)

function onVisibilityOrOnline() {
  if (!userStore.accessToken) return
  if (document.visibilityState === 'hidden') return
  chatStore.ensureAlive()
}

onMounted(() => {
  if (userStore.accessToken) {
    chatStore.connect()
  }
  document.addEventListener('visibilitychange', onVisibilityOrOnline)
  window.addEventListener('online', onVisibilityOrOnline)
})

onUnmounted(() => {
  document.removeEventListener('visibilitychange', onVisibilityOrOnline)
  window.removeEventListener('online', onVisibilityOrOnline)
})

function onChange(index: number) {
  const tab = tabs[index]
  if (tab) router.replace({ name: tab.routeName })
}
</script>

<template>
  <div class="layout">
    <div class="content">
      <router-view v-slot="{ Component }">
        <keep-alive>
          <component :is="Component" />
        </keep-alive>
      </router-view>
    </div>
    <van-tabbar :model-value="active" fixed safe-area-inset-bottom @change="onChange">
      <van-tabbar-item
        v-for="tab in tabs"
        :key="tab.routeName"
        :badge="tab.routeName === 'ConversationList' ? chatStore.tabUnreadBadge : undefined"
      >
        <template #icon>
          <TabBarIcons :name="tab.icon" class="tab-icon" />
        </template>
        {{ tab.label }}
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<style scoped>
.layout {
  position: fixed;
  top: var(--vv-offset-top, 0px);
  left: 0;
  right: 0;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: var(--app-vh, 100dvh);
  max-height: var(--app-vh, 100dvh);
}
.content {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  padding-bottom: var(--app-tabbar-height);
  display: flex;
  flex-direction: column;
  background: var(--page-bg);
}
.tab-icon {
  display: block;
  width: 26px;
  height: 26px;
}
:deep(.van-tabbar) {
  z-index: 100;
  height: 56px;
  --van-tabbar-height: 56px;
  --van-tabbar-item-icon-size: 26px;
  --van-tabbar-item-text-color: var(--text-primary);
  --van-tabbar-item-active-color: var(--brand-color);
  --van-tabbar-item-icon-margin-bottom: 3px;
  --van-tabbar-item-font-size: 12px;
  border-top: 1px solid var(--border-color);
  box-shadow: none;
}
</style>
