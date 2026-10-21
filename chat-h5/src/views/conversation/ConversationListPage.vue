<script setup lang="ts">
import { onActivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog, showSuccessToast } from 'vant'

import UserAvatar from '@/components/UserAvatar.vue'
import { pinConversation, type ConversationItem } from '@/api/conversation'
import { useChatStore } from '@/stores/chat'
import { sortConversations, useConversationList } from '@/composables/useConversationList'
import { useLongPress } from '@/composables/useLongPress'
import { formatConversationTime } from '@/utils/chat/chatTime'
import { openUserProfile, peerDisplayName } from '@/utils/user'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const chatStore = useChatStore()
const {
  list,
  loading,
  refreshing,
  finished,
  loadMore,
  refresh: onRefresh,
  reset: resetAndLoad,
} = useConversationList({
  live: true,
  onError: (e, firstPage) => showCatchError(e, '加载失败', firstPage),
})
const menuOpen = ref(false)
const menuStyle = ref<Record<string, string>>({})
const actionTarget = ref<ConversationItem | null>(null)
let hasLoaded = false
let pressTarget: ConversationItem | null = null
let menuOpenedAt = 0
let lastPressX = 0
let lastPressY = 0

const MENU_W = 132
const MENU_H = 38
const MENU_GAP = 8
const MENU_MARGIN = 8

function placeMenu(x: number, y: number) {
  const maxLeft = Math.max(MENU_MARGIN, window.innerWidth - MENU_W - MENU_MARGIN)
  const left = Math.min(Math.max(MENU_MARGIN, x - MENU_W / 2), maxLeft)
  let top = y - MENU_H - MENU_GAP
  if (top < MENU_MARGIN) {
    top = Math.min(y + MENU_GAP, window.innerHeight - MENU_H - MENU_MARGIN)
  }
  menuStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
  }
}

function openActions(c: ConversationItem, x: number, y: number) {
  actionTarget.value = c
  menuOpenedAt = Date.now()
  placeMenu(x, y)
  menuOpen.value = true
}

function closeMenu() {
  if (Date.now() - menuOpenedAt < 400) return
  menuOpen.value = false
  actionTarget.value = null
}

onMounted(async () => {
  await resetAndLoad()
  hasLoaded = true
})

onActivated(() => {
  if (!hasLoaded) return
  void chatStore.refreshUnreadSummary()
  void resetAndLoad()
})

function openChat(c: ConversationItem) {
  router.push({
    name: 'PrivateChat',
    params: { peerId: String(c.peerId) },
    query: { from: 'messages', name: peerDisplayName(c) },
  })
}

function openProfile(c: ConversationItem) {
  openUserProfile(router, c.peerId, {
    from: 'messages',
    name: peerDisplayName(c),
    avatarUrl: c.peerAvatar || '',
  })
}

const { onTouchStart, onTouchEnd, onTouchMove, consumeClickIfLongPressed } = useLongPress(() => {
  if (pressTarget) openActions(pressTarget, lastPressX, lastPressY)
})

function onRowTouchStart(e: TouchEvent, c: ConversationItem) {
  const t = e.touches[0]
  if (t) {
    lastPressX = t.clientX
    lastPressY = t.clientY
  }
  pressTarget = c
  onTouchStart(e)
}

function onRowClick(e: Event, c: ConversationItem) {
  if (consumeClickIfLongPressed(e)) return
  openChat(c)
}

function onContextMenu(e: MouseEvent, c: ConversationItem) {
  openActions(c, e.clientX, e.clientY)
}

function runMenuAction(action: 'pin' | 'delete') {
  const c = actionTarget.value
  menuOpen.value = false
  actionTarget.value = null
  if (!c) return
  if (action === 'pin') {
    void onTogglePin(c)
    return
  }
  void onDelete(c)
}

async function onTogglePin(c: ConversationItem) {
  const next = !c.pinned
  const prev = list.value
  list.value = sortConversations(
    list.value.map((item) => (item.peerId === c.peerId ? { ...item, pinned: next } : item)),
  )
  try {
    await pinConversation(c.peerId, next)
    showSuccessToast(next ? '已置顶' : '已取消置顶')
  } catch (e: unknown) {
    list.value = prev
    showCatchError(e, next ? '置顶失败' : '取消置顶失败')
  }
}

async function onDelete(c: ConversationItem) {
  try {
    await showConfirmDialog({
      title: `删除与「${peerDisplayName(c)}」的聊天`,
      message: '聊天记录将被清空，且无法恢复。',
      confirmButtonText: '删除',
      confirmButtonColor: 'var(--text-danger)',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  const peerId = c.peerId
  list.value = list.value.filter((item) => item.peerId !== peerId)

  try {
    await chatStore.deleteConversation(peerId)
    await resetAndLoad()
    showSuccessToast('已删除')
  } catch (e: unknown) {
    await resetAndLoad()
    showCatchError(e, '删除失败')
  }
}
</script>

<template>
  <div class="tab-page">
    <div class="tab-page__toolbar">
      <div class="tab-page__title">消息</div>
    </div>

    <van-pull-refresh v-model="refreshing" class="tab-page__scroll" @refresh="onRefresh">
      <div v-if="loading && list.length === 0 && !refreshing" class="tab-page__status">
        <van-loading size="24px" vertical>加载中…</van-loading>
      </div>
      <template v-else>
        <van-empty v-if="list.length === 0" description="暂无消息" />
        <van-list
          v-if="list.length > 0"
          v-model:loading="loading"
          :finished="finished"
          :immediate-check="false"
          finished-text="没有更多了"
          @load="loadMore"
        >
          <div class="msg-list">
            <van-cell
              v-for="c in list"
              :key="c.peerId"
              clickable
              class="conv-item"
              @click="onRowClick($event, c)"
              @touchstart.passive="onRowTouchStart($event, c)"
              @touchend="onTouchEnd"
              @touchmove="onTouchMove"
              @touchcancel="onTouchEnd"
              @contextmenu.prevent="onContextMenu($event, c)"
            >
              <template #icon>
                <div class="conv-avatar" @click.stop="openProfile(c)">
                  <UserAvatar :src="c.peerAvatar" :size="44" />
                </div>
              </template>
              <template #title>
                <div class="conv-title">
                  <span class="name">{{ peerDisplayName(c) }}</span>
                  <span class="meta">
                    <svg
                      v-if="c.pinned"
                      class="pin-icon"
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      aria-hidden="true"
                    >
                      <path
                        fill="currentColor"
                        d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"
                      />
                    </svg>
                    <span class="time">{{ formatConversationTime(c.lastActiveAt) }}</span>
                  </span>
                </div>
              </template>
              <template #label>
                <div class="conv-preview">
                  <span class="preview">{{ c.lastContent || '开始聊天吧' }}</span>
                  <van-badge
                    v-if="chatStore.getUnread(c.peerId) > 0"
                    :content="chatStore.getUnread(c.peerId)"
                  />
                </div>
              </template>
            </van-cell>
          </div>
        </van-list>
      </template>
    </van-pull-refresh>

    <Teleport to="body">
      <div
        v-if="menuOpen"
        class="conv-menu-mask"
        aria-hidden="true"
        @click="closeMenu"
        @touchstart.passive="closeMenu"
      />
      <div
        v-if="menuOpen && actionTarget"
        class="conv-menu"
        role="menu"
        :style="menuStyle"
        @click.stop
        @touchstart.stop
      >
        <button type="button" class="conv-menu__item" role="menuitem" @click="runMenuAction('pin')">
          {{ actionTarget.pinned ? '取消置顶' : '置顶' }}
        </button>
        <button
          type="button"
          class="conv-menu__item"
          role="menuitem"
          @click="runMenuAction('delete')"
        >
          删除
        </button>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.msg-list {
  background: transparent;
}

.conv-item {
  background: transparent;
}

.conv-avatar {
  margin-right: 12px;
}

.conv-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.conv-title .name {
  min-width: 0;
  font-size: 15px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-title .meta {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
}

.pin-icon {
  display: block;
  color: var(--text-tertiary);
}

.conv-title .time {
  font-size: 11px;
  color: var(--text-muted);
}

.conv-preview {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
}

.conv-preview .preview {
  color: #888888;
  font-size: 12px;
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-item :deep(.van-cell__value) {
  flex: none;
}

.conv-menu {
  position: fixed;
  z-index: 30;
  display: flex;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.7);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.16);
  overflow: hidden;
}

.conv-menu__item {
  padding: 9px 14px;
  border: none;
  background: transparent;
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.2;
  white-space: nowrap;
  cursor: pointer;
}

.conv-menu-mask {
  position: fixed;
  inset: 0;
  z-index: 25;
}
</style>
