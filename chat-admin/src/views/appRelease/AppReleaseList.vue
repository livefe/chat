<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <AppReleaseFilter v-model:platform="filterPlatform" @search="loadList" />
          </div>
        </div>
      </template>

      <section
        v-if="can(PERM.OpsAppReleasePolicy)"
        class="release-section"
        v-loading="policyLoading"
      >
        <AppReleasePolicyPanel :policies="policies" @saved="loadPolicies" />
      </section>

      <section v-if="canShowPublishCard" class="release-section">
        <AppReleasePublishForm @published="loadList" />
      </section>

      <section class="release-section release-section--last">
        <div class="section-title">版本列表</div>
        <AppReleaseTable
          :list="list"
          :loading="loading"
          @enable="handleEnable"
          @disable="handleDisable"
        />
      </section>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  disableAppRelease,
  enableAppRelease,
  getAppReleaseList,
  getAppUpdatePolicyList,
  type AppReleaseItem,
  type AppUpdatePolicyItem,
} from '@/api/appRelease'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'
import AppReleaseFilter from './AppReleaseFilter.vue'
import AppReleasePolicyPanel from './AppReleasePolicyPanel.vue'
import AppReleasePublishForm from './AppReleasePublishForm.vue'
import AppReleaseTable from './AppReleaseTable.vue'

const route = useRoute()
const { can } = usePermission()

const canShowPublishCard = computed(
  () => can(PERM.OpsAppReleaseUpload) || can(PERM.OpsAppReleasePublish),
)

const list = ref<AppReleaseItem[]>([])
const policies = ref<AppUpdatePolicyItem[]>([])
const loading = ref(false)
const policyLoading = ref(false)
const filterPlatform = ref('')

async function loadList() {
  loading.value = true
  try {
    const res = await getAppReleaseList(filterPlatform.value || undefined)
    list.value = res.data.list || []
  } catch (e) {
    showApiError(e, '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadPolicies() {
  policyLoading.value = true
  try {
    const res = await getAppUpdatePolicyList()
    policies.value = res.data.list || []
  } catch (e) {
    showApiError(e, '加载更新策略失败')
  } finally {
    policyLoading.value = false
  }
}

async function handleEnable(row: AppReleaseItem) {
  if (row.isCurrent === 1) return
  try {
    await ElMessageBox.confirm(
      `确认将 ${row.platformText} ${row.version} 设为当前对外版本？用户将收到更新提示（不会自动强制）。`,
      '设为当前',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await enableAppRelease(row.id)
    ElMessage.success('已更新')
    await Promise.all([loadList(), loadPolicies()])
  } catch (e) {
    showApiError(e, '操作失败')
  }
}

async function handleDisable(row: AppReleaseItem) {
  if (row.isCurrent !== 1) return
  try {
    await ElMessageBox.confirm(
      `确认下架 ${row.platformText}？下架后将不再对外展示该平台下载入口。`,
      '下架',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await disableAppRelease(row.id)
    ElMessage.success('已下架')
    await Promise.all([loadList(), loadPolicies()])
  } catch (e) {
    showApiError(e, '操作失败')
  }
}

onMounted(() => Promise.all([loadList(), loadPolicies()]))
</script>

<style scoped>
.release-section {
  padding-bottom: 24px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.release-section--last {
  padding-bottom: 0;
  margin-bottom: 0;
  border-bottom: none;
}

.release-section > .section-title {
  margin-bottom: 12px;
}
</style>
