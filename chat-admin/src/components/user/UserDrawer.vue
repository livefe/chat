<template>
  <el-drawer v-model="drawer.visible" title="用户详情" size="880px" destroy-on-close>
    <div v-loading="loading" class="user-drawer">
      <template v-if="detail">
        <div class="user-drawer__head">
          <el-avatar :src="detail.avatarUrl || undefined" :size="56">
            {{ (detail.nickname || '用').slice(0, 1) }}
          </el-avatar>
          <div class="user-drawer__meta">
            <div class="user-drawer__name">
              <span>{{ detail.nickname || '未设置昵称' }}</span>
              <el-tag v-if="detail.deactivated" type="info" size="small">已注销</el-tag>
              <el-tag v-else :type="detail.status === 1 ? 'success' : 'danger'" size="small">
                {{ detail.statusText }}
              </el-tag>
              <el-tag v-if="detail.muteText" type="warning" size="small">
                {{ detail.muteText }}
              </el-tag>
            </div>
            <div class="sub">
              ID <CopyText :text="detail.id" />
              <span class="user-drawer__dot">·</span>
              <template v-if="fullMobile">
                <CopyText :text="fullMobile" />
              </template>
              <template v-else>
                {{ detail.mobile || '号码已清除' }}
                <el-button link type="primary" size="small" :loading="revealing" @click="reveal">
                  {{ detail.deactivated ? '查看原手机号' : '查看完整手机号' }}
                </el-button>
              </template>
            </div>
            <div class="sub">
              注册 {{ detail.createdAt }}
              <span class="user-drawer__dot">·</span>
              最近登录 {{ detail.lastLoginAt || '-' }}
            </div>
          </div>
        </div>

        <div class="user-drawer__actions">
          <template v-if="!detail.deactivated">
            <el-button
              v-if="can(PERM.UserAccountUpdate)"
              size="small"
              @click="actions.openEdit(target)"
            >
              编辑
            </el-button>
            <template v-if="can(PERM.UserAccountMute)">
              <el-button size="small" type="warning" plain @click="actions.openMute([target])">
                {{ detail.muteText ? '调整禁言' : '禁言' }}
              </el-button>
              <el-button
                v-if="detail.muteText"
                size="small"
                type="warning"
                plain
                @click="actions.unmute(target)"
              >
                解除禁言
              </el-button>
            </template>
            <template v-if="can(PERM.UserAccountStatus)">
              <el-button
                v-if="detail.status === UserStatus.Enabled"
                size="small"
                type="danger"
                plain
                @click="actions.disable(target)"
              >
                禁用
              </el-button>
              <el-button v-else size="small" type="success" plain @click="actions.enable(target)">
                启用
              </el-button>
            </template>
            <el-button
              v-if="can(PERM.UserAccountDeactivate)"
              size="small"
              type="danger"
              @click="actions.deactivate(target)"
            >
              注销
            </el-button>
          </template>
          <el-button
            v-if="can(PERM.ChatMessage)"
            class="user-drawer__all-messages"
            size="small"
            link
            type="primary"
            @click="go(PERM.ChatMessage, { userId: String(detail.id) })"
          >
            查看全部消息
          </el-button>
        </div>

        <el-tabs :key="drawer.version" v-model="drawer.tab" class="user-drawer__tabs">
          <el-tab-pane label="概览" name="overview" lazy>
            <UserOverviewTab :user-id="detail.id" :detail="detail" />
          </el-tab-pane>
          <el-tab-pane label="处罚记录" name="punishments" lazy>
            <UserPunishmentsTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane v-if="can(PERM.PostPost)" label="帖子" name="posts" lazy>
            <UserPostsTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane v-if="can(PERM.PostPost)" label="报名" name="signups" lazy>
            <UserSignupsTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane v-if="can(PERM.ChatConversation)" label="私信会话" name="conversations" lazy>
            <UserConversationsTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane v-if="can(PERM.UserReport)" label="举报" name="reports" lazy>
            <UserReportsTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane v-if="can(PERM.UserFeedback)" label="反馈" name="feedback" lazy>
            <UserFeedbackTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane label="设备" name="devices" lazy>
            <UserDevicesTab :user-id="detail.id" />
          </el-tab-pane>
          <el-tab-pane v-if="can(PERM.ImSupportOverview)" label="客服会话" name="support" lazy>
            <UserSupportTab :user-id="detail.id" />
          </el-tab-pane>
        </el-tabs>
      </template>
    </div>
  </el-drawer>
  <UserMuteDialog />
  <UserFormDialog />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { getUserDetail, revealUserMobile, type AdminUserDetail } from '@/api/user'
import CopyText from '@/components/common/CopyText.vue'
import { usePermission } from '@/composables/usePermission'
import { useUserActions, type UserActionTarget } from '@/composables/useUserActions'
import { PERM } from '@/constants/permission'
import { UserStatus } from '@/constants/user'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { showApiError } from '@/utils/apiError'
import UserConversationsTab from './tabs/UserConversationsTab.vue'
import UserDevicesTab from './tabs/UserDevicesTab.vue'
import UserFeedbackTab from './tabs/UserFeedbackTab.vue'
import UserOverviewTab from './tabs/UserOverviewTab.vue'
import UserPostsTab from './tabs/UserPostsTab.vue'
import UserPunishmentsTab from './tabs/UserPunishmentsTab.vue'
import UserReportsTab from './tabs/UserReportsTab.vue'
import UserSignupsTab from './tabs/UserSignupsTab.vue'
import UserSupportTab from './tabs/UserSupportTab.vue'
import { useDrawerNav } from './tabs/useDrawerNav'
import UserFormDialog from './UserFormDialog.vue'
import UserMuteDialog from './UserMuteDialog.vue'

const drawer = useUserDrawerStore()
const { can } = usePermission()
const { go } = useDrawerNav()
const actions = useUserActions()

const loading = ref(false)
const detail = ref<AdminUserDetail | null>(null)
const fullMobile = ref('')
const revealing = ref(false)

const target = computed<UserActionTarget>(() => ({
  id: detail.value?.id ?? 0,
  nickname: detail.value?.nickname,
  mobile: detail.value?.mobile,
  status: detail.value?.status,
}))

async function load() {
  const id = drawer.userId
  if (!id) return
  loading.value = true
  try {
    const { data } = await getUserDetail(id)
    if (id === drawer.userId) detail.value = data
  } catch (e) {
    showApiError(e, '获取用户详情失败')
    drawer.close()
  } finally {
    loading.value = false
  }
}

async function reveal() {
  if (!detail.value) return
  revealing.value = true
  try {
    const { data } = await revealUserMobile(detail.value.id)
    fullMobile.value = data.mobile
  } catch (e) {
    showApiError(e, '获取手机号失败')
  } finally {
    revealing.value = false
  }
}

watch(
  () => [drawer.visible, drawer.version] as const,
  ([visible]) => {
    if (!visible) return
    if (drawer.userId !== detail.value?.id) {
      detail.value = null
      fullMobile.value = ''
    }
    void load()
  },
)
</script>

<style scoped>
.user-drawer {
  min-height: 200px;
}

.user-drawer__head {
  display: flex;
  gap: 14px;
  align-items: center;
}

.user-drawer__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.user-drawer__name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.user-drawer__dot {
  margin: 0 6px;
}

.user-drawer__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin: 14px 0 6px;
}

.user-drawer__actions .el-button + .el-button {
  margin-left: 0;
}

.user-drawer__all-messages {
  margin-left: auto !important;
}
</style>

<style>
.user-tab__section {
  margin: 16px 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.user-tab__toolbar {
  margin-bottom: 10px;
}

.user-tab__pager {
  justify-content: flex-end;
  margin-top: 10px;
}

.user-tab__row {
  cursor: pointer;
}
</style>
