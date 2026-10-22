<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
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
        <AppReleasePublishForm :policies="policies" @published="onPublished" />
      </section>

      <section class="release-section release-section--last">
        <div class="section-head">
          <div class="section-title">版本列表</div>
          <AppReleaseFilter v-model:platform="query.platform" @search="handleSearch" />
        </div>
        <AppReleaseTable
          :list="list"
          :loading="loading"
          :error="error"
          :total="total"
          v-model:page="query.page"
          v-model:page-size="query.pageSize"
          @enable="handleEnable"
          @disable="handleDisable"
          @delete="handleDelete"
          @page-change="fetchList"
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
  deleteAppRelease,
  disableAppRelease,
  enableAppRelease,
  getAppReleaseList,
  getAppUpdatePolicyList,
  type AppReleaseItem,
  type AppUpdatePolicyItem,
} from '@/api/appRelease'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { compareSemver } from '@/constants/appRelease'
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

const { query, list, total, loading, error, fetchList, handleSearch, afterRemove } = useListQuery({
  fetch: (q) => getAppReleaseList(q),
  defaults: { page: 1, pageSize: 20, platform: '' },
  schema: { platform: 'string' },
  errorMessage: '获取版本列表失败',
})

const policies = ref<AppUpdatePolicyItem[]>([])
const policyLoading = ref(false)

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

function onPublished() {
  void handleSearch()
}

function enableMessage(row: AppReleaseItem) {
  const policy = policies.value.find((p) => p.platform === row.platform)
  const current = policy?.latestVersion || ''
  const minVer = policy?.minSupportedVersion || ''
  const lines: string[] = []
  if (current && compareSemver(row.version, current) < 0) {
    lines.push(`这是一次回滚：对外版本将从 ${current} 降为 ${row.version}。`)
  } else {
    lines.push(
      `确认将 ${row.platformText} ${row.version} 设为当前对外版本？低于该版本的用户将收到更新提示（不强制）。`,
    )
  }
  if (minVer && compareSemver(row.version, minVer) < 0) {
    lines.push(
      `最低支持版本将从 ${minVer} 降到 ${row.version}，版本在 ${row.version} 与 ${minVer} 之间的客户端不再被强制更新。`,
    )
  }
  return lines.join('\n')
}

async function handleEnable(row: AppReleaseItem) {
  if (row.isCurrent === 1) return
  try {
    await ElMessageBox.confirm(enableMessage(row), '设为当前版本', {
      type: 'warning',
      customStyle: { whiteSpace: 'pre-line' },
    })
  } catch {
    return
  }
  try {
    await enableAppRelease(row.id)
    ElMessage.success('已设为当前版本')
    await Promise.all([fetchList(), loadPolicies()])
  } catch (e) {
    showApiError(e, '操作失败')
  }
}

async function handleDisable(row: AppReleaseItem) {
  if (row.isCurrent !== 1) return
  try {
    await ElMessageBox.confirm(
      `确认下架 ${row.platformText} ${row.version}？下架后该平台不再对外展示下载入口，也不会推送更新提示。`,
      '下架',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await disableAppRelease(row.id)
    ElMessage.success('已下架')
    await Promise.all([fetchList(), loadPolicies()])
  } catch (e) {
    showApiError(e, '操作失败')
  }
}

async function handleDelete(row: AppReleaseItem) {
  if (row.isCurrent === 1) return
  const fileTip = row.platform === 'h5' ? '' : '安装包文件将一并删除，'
  try {
    await ElMessageBox.confirm(
      `确认删除 ${row.platformText} ${row.version}？${fileTip}此操作不可恢复。`,
      '删除版本',
      { type: 'warning', confirmButtonText: '删除' },
    )
  } catch {
    return
  }
  try {
    await deleteAppRelease(row.id)
    ElMessage.success('已删除')
    await afterRemove()
  } catch (e) {
    showApiError(e, '删除失败')
  }
}

onMounted(loadPolicies)
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

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}
</style>
