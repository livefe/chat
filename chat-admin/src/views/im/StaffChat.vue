<template>
  <div class="admin-page im-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <ImAttentionToggle />
          </div>
        </div>
      </template>

      <ImConnectionBanner :state="wsState" @reconnect="reconnect" />

      <ImShell v-model:active-tab="sideTab" :tabs="tabs">
        <template #search>
          <el-input
            v-if="sideTab === 'sessions'"
            v-model="sessionKeyword"
            placeholder="搜索会话（同事姓名）"
            clearable
            @keyup.enter="loadSessions()"
            @clear="loadSessions()"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-input
            v-else
            v-model="contactKeyword"
            placeholder="搜索同事姓名或手机号"
            clearable
            @keyup.enter="loadContacts()"
            @clear="loadContacts()"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </template>

        <template #list>
          <div v-if="sideTab === 'sessions'" v-loading="sessionsLoading">
            <ImListItem
              v-for="s in sessions"
              :key="s.id"
              :title="s.peerName || s.title || '同事'"
              :preview="s.lastContent || '暂无消息'"
              :time="imFormatTime(s.lastActiveAt)"
              :avatar-url="s.peerAvatarUrl"
              :online="s.peerOnline"
              :unread="s.unreadCount"
              :active="active?.id === s.id"
              @click="selectSession(s)"
            />
            <div v-if="sessions.length < sessionsTotal" class="im-sidebar-more">
              <el-button
                link
                type="primary"
                :loading="sessionsLoadingMore"
                @click="loadSessions({ append: true })"
              >
                加载更多
              </el-button>
            </div>
            <div v-if="!sessionsLoading && sessions.length === 0" class="im-sidebar-empty">
              <el-empty
                :description="sessionKeyword ? '没有匹配的会话' : '暂无会话，去通讯录发起'"
                :image-size="72"
              />
            </div>
          </div>
          <div v-else v-loading="contactsLoading">
            <ImListItem
              v-for="c in contacts"
              :key="c.id"
              :title="c.name || String(c.id)"
              :subtitle="c.departmentName || '未分配部门'"
              :avatar-url="c.avatarUrl"
              :online="c.online"
              :disabled="!canOpen"
              @click="startChat(c)"
            />
            <div v-if="contacts.length < contactsTotal" class="im-sidebar-more">
              <el-button
                link
                type="primary"
                :loading="contactsLoadingMore"
                @click="loadContacts({ append: true })"
              >
                加载更多
              </el-button>
            </div>
            <div v-if="!contactsLoading && contacts.length === 0" class="im-sidebar-empty">
              <el-empty description="暂无同事" :image-size="72" />
            </div>
          </div>
        </template>

        <template #main>
          <template v-if="active">
            <ImChatHeader :title="active.peerName || active.title">
              <template #tags>
                <el-tag size="small" :type="active.peerOnline ? 'success' : 'info'" effect="plain">
                  {{ active.peerOnline ? '在线' : '离线' }}
                </el-tag>
              </template>
              <template #sub>{{ active.peerDepartmentName || '未分配部门' }}</template>
            </ImChatHeader>
            <ImMessageThread
              ref="threadRef"
              :messages="messages"
              :self-id="selfId"
              mine-mode="staff"
              :loading="messagesLoading"
              :has-more="hasMore"
              :loading-more="loadingMore"
              :show-recall="canRecall"
              @load-more="loadMore"
              @recall="onRecall"
              @resend="onResend"
            />
            <ImComposer
              v-if="canSend"
              v-model="draft"
              :sending="sending"
              :uploading="uploading"
              :allow-upload="canUpload"
              @send="onSend"
              @upload="onUpload"
            />
            <div v-else class="im-closed-bar">无发送权限</div>
          </template>
          <div v-else class="im-empty">
            <div class="im-empty__title">{{ route.meta.title }}</div>
            <div class="im-empty__desc">从左侧会话列表继续聊天，或在通讯录中选择同事</div>
            <el-button type="primary" @click="sideTab = 'contacts'">打开通讯录</el-button>
          </div>
        </template>
      </ImShell>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { useStaffChat } from '@/composables/useStaffChat'
import ImAttentionToggle from '@/components/im/ImAttentionToggle.vue'
import ImConnectionBanner from '@/components/im/ImConnectionBanner.vue'
import ImShell from '@/components/im/ImShell.vue'
import ImListItem from '@/components/im/ImListItem.vue'
import ImChatHeader from '@/components/im/ImChatHeader.vue'
import ImMessageThread from '@/components/im/ImMessageThread.vue'
import ImComposer from '@/components/im/ImComposer.vue'
import { imFormatTime } from '@/components/im/media'

const route = useRoute()
const {
  selfId,
  canOpen,
  canSend,
  canRecall,
  canUpload,
  sideTab,
  tabs,
  sessionKeyword,
  sessions,
  sessionsTotal,
  sessionsLoading,
  sessionsLoadingMore,
  contactKeyword,
  contacts,
  contactsTotal,
  contactsLoading,
  contactsLoadingMore,
  active,
  messages,
  messagesLoading,
  loadingMore,
  hasMore,
  draft,
  sending,
  uploading,
  threadRef,
  wsState,
  reconnect,
  loadContacts,
  loadSessions,
  startChat,
  selectSession,
  loadMore,
  onSend,
  onResend,
  onUpload,
  onRecall,
} = useStaffChat()
</script>

<style scoped>
.im-sidebar-more {
  display: flex;
  justify-content: center;
  padding: 8px 0 12px;
}
</style>
