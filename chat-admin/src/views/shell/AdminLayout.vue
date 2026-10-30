<template>
  <el-container class="layout-container">
    <el-aside :width="sidebarCollapsed ? '64px' : '210px'" class="sidebar">
      <div class="logo">{{ sidebarCollapsed ? '管' : '管理后台' }}</div>
      <div v-if="!sidebarCollapsed" class="menu-search">
        <el-input
          ref="searchRef"
          v-model="menuKeyword"
          size="small"
          clearable
          placeholder="搜索菜单（Ctrl+K）"
          :prefix-icon="Search"
          @keydown.enter.prevent="openFirstMatch"
          @keydown.esc="menuKeyword = ''"
        />
      </div>
      <el-menu
        :key="menuRenderKey"
        :default-active="route.path"
        :default-openeds="searchOpeneds"
        class="sidebar-menu"
        :collapse="sidebarCollapsed"
        :collapse-transition="false"
        :background-color="menuColors.bg"
        :text-color="menuColors.text"
        :active-text-color="menuColors.active"
        @select="onMenuSelect"
      >
        <el-menu-item v-if="!searching" :index="homePath">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>
        <template v-for="menu in displayMenus" :key="menu.path || menu.title">
          <el-menu-item v-if="!menu.children || menu.children.length === 0" :index="menu.path">
            <el-icon v-if="menu.icon"><component :is="menu.icon" /></el-icon>
            <span>{{ menu.title }}</span>
            <el-badge
              v-if="badgeFor(menu.path)"
              :value="badgeFor(menu.path)"
              :max="99"
              class="menu-badge"
            />
          </el-menu-item>
          <el-sub-menu v-else :index="submenuIndex(menu)">
            <template #title>
              <el-icon v-if="menu.icon"><component :is="menu.icon" /></el-icon>
              <span>{{ menu.title }}</span>
            </template>
            <el-menu-item v-for="child in menu.children" :key="child.path" :index="child.path">
              <span>{{ child.title }}</span>
              <el-badge
                v-if="badgeFor(child.path)"
                :value="badgeFor(child.path)"
                :max="99"
                class="menu-badge"
              />
            </el-menu-item>
          </el-sub-menu>
        </template>
        <div v-if="searching && displayMenus.length === 0" class="menu-search-empty">
          无匹配菜单
        </div>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-button class="sidebar-toggle" text @click="toggleSidebar">
            <el-icon :size="18">
              <Fold v-if="!sidebarCollapsed" />
              <Expand v-else />
            </el-icon>
          </el-button>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: homePath }">首页</el-breadcrumb-item>
            <el-breadcrumb-item
              v-for="(item, index) in crumbTrail"
              :key="item.path || item.title"
              :to="
                index < crumbTrail.length - 1 && isRoutable(item.path)
                  ? { path: item.path }
                  : undefined
              "
            >
              {{ item.title }}
            </el-breadcrumb-item>
            <el-breadcrumb-item
              v-if="route.path !== homePath && crumbTrail.length === 0 && route.meta.title"
            >
              {{ route.meta.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="28" :src="employeeStore.employeeInfo?.avatarUrl || undefined">
              {{
                (
                  employeeStore.employeeInfo?.name ||
                  employeeStore.employeeInfo?.mobile ||
                  '管'
                ).slice(0, 1)
              }}
            </el-avatar>
            {{ employeeStore.employeeInfo?.name || employeeStore.employeeInfo?.mobile }}
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>

    <UserDrawer />

    <el-drawer v-model="profileVisible" title="个人资料" size="420px">
      <div v-loading="profileLoading" class="profile-drawer">
        <el-form v-if="profileForm" label-width="72px" @submit.prevent>
          <el-form-item label="头像">
            <div class="profile-avatar">
              <el-avatar :size="64" :src="profileForm.avatarUrl || undefined">
                {{ (profileForm.name || profileForm.mobile || '管').slice(0, 1) }}
              </el-avatar>
              <div class="profile-avatar__ops">
                <el-upload
                  :show-file-list="false"
                  accept="image/*"
                  :http-request="onUploadAvatar"
                  :disabled="uploading || saving"
                >
                  <el-button :loading="uploading" :disabled="saving">上传头像</el-button>
                </el-upload>
                <el-button
                  v-if="profileForm.avatarUrl"
                  link
                  type="danger"
                  :disabled="saving"
                  @click="clearAvatar"
                >
                  清除
                </el-button>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="姓名">
            <el-input :model-value="profileForm.name || '-'" disabled />
          </el-form-item>
          <el-form-item label="手机号">
            <el-input :model-value="profileForm.mobile" disabled />
          </el-form-item>
          <el-form-item label="工号">
            <el-input :model-value="profileForm.employeeNo || '-'" disabled />
          </el-form-item>
          <el-form-item label="部门">
            <el-input :model-value="profileForm.department || '-'" disabled />
          </el-form-item>
          <el-form-item label="状态">
            <el-tag :type="profileForm.status === 1 ? 'success' : 'danger'">
              {{ profileForm.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="saveProfile">保存</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-drawer>
  </el-container>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ElMessage,
  ElMessageBox,
  type InputInstance,
  type UploadRequestHandler,
} from 'element-plus'
import { Expand, Fold, HomeFilled, Search } from '@element-plus/icons-vue'
import { showApiError } from '@/utils/apiError'
import { useEmployeeStore } from '@/stores/employee'
import { getEmployeeMe, updateMyProfile, uploadMyAvatar, type EmployeeInfo } from '@/api/employee'
import { useAdminBadges } from '@/composables/useAdminBadges'
import {
  buildMenus,
  filterMenus,
  findMenuTrail,
  flattenMenuLeaves,
  type MenuItem,
} from '@/utils/menu'
import { pushRecentVisit } from '@/utils/storage'
import { HOME_PATH } from '@/router/dynamic'
import UserDrawer from '@/components/user/UserDrawer.vue'

const SIDEBAR_COLLAPSED_KEY = 'admin_sidebar_collapsed'

const route = useRoute()
const router = useRouter()
const employeeStore = useEmployeeStore()

const menus = computed(() => buildMenus(employeeStore.permissions))
const homePath = HOME_PATH

const { badgeFor, pendingQueryFor } = useAdminBadges()

const searchRef = ref<InputInstance>()
const menuKeyword = ref('')
const searching = computed(() => menuKeyword.value.trim() !== '')
const displayMenus = computed(() => filterMenus(menus.value, menuKeyword.value))
const menuRenderKey = computed(() => (searching.value ? `search:${menuKeyword.value}` : 'all'))

function submenuIndex(menu: MenuItem) {
  return `submenu:${menu.path || menu.title}`
}

const searchOpeneds = computed(() =>
  searching.value ? displayMenus.value.filter((m) => m.children?.length).map(submenuIndex) : [],
)

function navigateMenu(path: string) {
  menuKeyword.value = ''
  void router.push({ path, query: pendingQueryFor(path) })
}

function onMenuSelect(index: string) {
  if (index.startsWith('submenu:')) return
  navigateMenu(index)
}

function openFirstMatch() {
  if (!searching.value) return
  const first = flattenMenuLeaves(displayMenus.value)[0]
  if (first) navigateMenu(first.path)
}

function onGlobalKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    sidebarCollapsed.value = false
    void nextTick(() => searchRef.value?.focus())
  }
}

const crumbTrail = computed(() =>
  route.path === homePath ? [] : findMenuTrail(menus.value, route.path),
)

function isRoutable(path: string) {
  if (!path) return false
  const resolved = router.resolve(path)
  return resolved.matched.length > 0 && resolved.name !== 'NotFound'
}

watch(
  () => route.path,
  () => {
    const employeeId = employeeStore.employeeInfo?.id
    const title = route.meta.title
    if (!employeeId || !route.meta.permission || typeof title !== 'string') return
    pushRecentVisit(employeeId, { path: route.path, title })
  },
  { immediate: true },
)

const sidebarCollapsed = ref(localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === '1')

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem(SIDEBAR_COLLAPSED_KEY, sidebarCollapsed.value ? '1' : '0')
}

const menuColors = {
  bg: 'var(--admin-sidebar-bg)',
  text: 'var(--admin-sidebar-text)',
  active: 'var(--admin-sidebar-active)',
}

const profileVisible = ref(false)
const profileLoading = ref(false)
const uploading = ref(false)
const saving = ref(false)
const profileForm = reactive({
  name: '',
  mobile: '',
  employeeNo: '',
  avatarMediaId: 0,
  avatarUrl: '',
  department: '',
  status: 1,
})

function applyProfile(data: EmployeeInfo) {
  profileForm.name = data.name || ''
  profileForm.mobile = data.mobile || ''
  profileForm.employeeNo = data.employeeNo || ''
  profileForm.avatarMediaId = data.avatarMediaId || 0
  profileForm.avatarUrl = data.avatarUrl || ''
  profileForm.department = data.department || ''
  profileForm.status = data.status
}

async function refreshPermissions() {
  if (!employeeStore.isAuthenticated) return
  try {
    await employeeStore.loadPermissions()
  } catch {
    ElMessage.warning('权限刷新失败，请稍后重试')
  }
}

function onFocus() {
  void refreshPermissions()
}

onMounted(() => {
  window.addEventListener('focus', onFocus)
  window.addEventListener('keydown', onGlobalKeydown)
  if (window.innerWidth < 960 && !sidebarCollapsed.value) {
    sidebarCollapsed.value = true
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('focus', onFocus)
  window.removeEventListener('keydown', onGlobalKeydown)
})

async function openProfile() {
  profileVisible.value = true
  profileLoading.value = true
  try {
    const { data } = await getEmployeeMe()
    applyProfile(data)
    employeeStore.setEmployeeInfo(data)
  } catch (e) {
    showApiError(e, '获取个人资料失败')
    profileVisible.value = false
  } finally {
    profileLoading.value = false
  }
}

const onUploadAvatar: UploadRequestHandler = async (options) => {
  uploading.value = true
  try {
    const { data } = await uploadMyAvatar(options.file as File)
    profileForm.avatarMediaId = data.id
    profileForm.avatarUrl = data.url || ''
    options.onSuccess?.(data)
  } catch (e) {
    options.onError?.(e as Parameters<NonNullable<typeof options.onError>>[0])
    showApiError(e, '上传头像失败')
  } finally {
    uploading.value = false
  }
}

function clearAvatar() {
  profileForm.avatarMediaId = 0
  profileForm.avatarUrl = ''
}

async function saveProfile() {
  saving.value = true
  try {
    const { data } = await updateMyProfile({
      avatarMediaId: profileForm.avatarMediaId,
    })
    applyProfile(data)
    employeeStore.setEmployeeInfo(data)
    ElMessage.success('保存成功')
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleCommand(command: string) {
  if (command === 'profile') {
    await openProfile()
    return
  }
  if (command !== 'logout') return
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  employeeStore.logout()
  ElMessage.success('退出成功')
  void router.push('/login')
}
</script>

<style scoped>
.header-left {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.sidebar-toggle {
  padding: 6px;
  margin-right: 4px;
  color: var(--el-text-color-regular);
}

.menu-search {
  padding: 10px 12px 4px;
  flex-shrink: 0;
}

.menu-search-empty {
  padding: 16px;
  font-size: 13px;
  text-align: center;
  color: var(--admin-sidebar-text);
}

.user-info {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.profile-avatar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.profile-avatar__ops {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sidebar-menu :deep(.el-menu-item) {
  position: relative;
}

.menu-badge {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  height: auto;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  pointer-events: none;
}

.menu-badge :deep(.el-badge__content) {
  position: static;
  transform: none;
  border: none;
}

.sidebar-menu.el-menu--collapse .menu-badge {
  right: 8px;
  top: 8px;
  transform: none;
}
</style>
