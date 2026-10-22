<template>
  <TreeManageLayout
    :title="String(route.meta.title || '行政区划')"
    tree-title="区划树"
    :main-loading="listLoading"
  >
    <template #search>
      <RegionFilter
        v-model:keyword="searchForm.keyword"
        v-model:level="searchForm.level"
        @search="handleSearch"
        @reset="handleResetSearch"
      />
    </template>

    <template #hits>
      <RegionSearchHits
        v-if="searchHits.length"
        :hits="searchHits"
        :loading="searchLoading"
        :truncated="searchTruncated"
        :limit="SEARCH_LIMIT"
        @locate="locateFromSearch"
        @close="searchHits = []"
      />
    </template>

    <template #tree>
      <div class="region-root-node" :class="{ 'is-active': !currentNode.code }" @click="goRoot">
        全国
      </div>
      <el-tree
        ref="treeRef"
        :props="treeProps"
        node-key="code"
        lazy
        highlight-current
        :expand-on-click-node="false"
        :load="loadTreeNode"
        @node-click="onTreeClick"
      />
    </template>

    <template #breadcrumb>
      <el-breadcrumb separator="/" class="tree-manage-crumb">
        <el-breadcrumb-item v-for="(item, index) in stack" :key="item.code || 'root'">
          <a href="javascript:;" @click="goCrumb(index)">{{ item.name }}</a>
        </el-breadcrumb-item>
      </el-breadcrumb>
    </template>

    <template #main-actions>
      <el-button
        v-if="canViewDetail && currentNode.code"
        size="small"
        @click="openDetail(currentNode.code)"
      >
        查看当前
      </el-button>
    </template>

    <template #main>
      <RegionChildTable
        :list="children"
        :can-view-detail="canViewDetail"
        :highlight-code="highlightCode"
        @enter="enterRegion"
        @detail="(row) => openDetail(row.code)"
      />
    </template>

    <RegionDetailDrawer
      v-model="detailVisible"
      :detail="detail"
      :loading="detailLoading"
      @enter="enterFromDetail"
    />
  </TreeManageLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type LoadFunction, type TreeInstance } from 'element-plus'
import {
  getRegionChildren,
  getRegionDetail,
  getRegionTrail,
  searchRegion,
  type RegionAdminItem,
  type RegionItem,
} from '@/api/region'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { REGION_LEAF_LEVEL } from '@/constants/region'
import { showApiError } from '@/utils/apiError'
import RegionChildTable from './RegionChildTable.vue'
import RegionDetailDrawer from './RegionDetailDrawer.vue'
import RegionFilter from './RegionFilter.vue'
import RegionSearchHits from './RegionSearchHits.vue'
import TreeManageLayout from './TreeManageLayout.vue'

type Crumb = { code: string; name: string; level: number }
type TreeNode = RegionItem & { isLeaf?: boolean }

const ROOT: Crumb = { code: '', name: '全国', level: 0 }
const SEARCH_LIMIT = 50

const { can } = usePermission()
const route = useRoute()
const router = useRouter()
const canViewDetail = computed(() => can(PERM.OpsRegionDetail))

const treeRef = ref<TreeInstance>()
const treeProps = {
  label: 'name',
  children: 'children',
  isLeaf: 'isLeaf',
}

const searchForm = reactive({ keyword: '', level: undefined as number | undefined })
const searchLoading = ref(false)
const searchHits = ref<RegionItem[]>([])
const searchTruncated = ref(false)

const stack = ref<Crumb[]>([ROOT])
const children = ref<RegionItem[]>([])
const highlightCode = ref('')
const listLoading = ref(false)
const detail = ref<RegionAdminItem | null>(null)
const detailLoading = ref(false)
const detailVisible = ref(false)

const currentNode = computed(() => stack.value[stack.value.length - 1] ?? ROOT)

function toCrumb(item: RegionItem): Crumb {
  return { code: item.code, name: item.name, level: item.level }
}

function toTreeNodes(list: RegionItem[]): TreeNode[] {
  return list.map((item) => ({
    ...item,
    isLeaf: item.level >= REGION_LEAF_LEVEL,
  }))
}

async function fetchChildren(parentCode: string): Promise<RegionItem[]> {
  const res = await getRegionChildren(parentCode)
  return res.data.list || []
}

const loadTreeNode: LoadFunction = async (node, resolve) => {
  try {
    const parentCode = node.level === 0 ? '' : (node.data as TreeNode).code
    const list = await fetchChildren(parentCode)
    resolve(toTreeNodes(list))
  } catch (e) {
    showApiError(e, '加载区划树失败')
    resolve([])
  }
}

async function loadChildren(parentCode: string) {
  listLoading.value = true
  try {
    children.value = await fetchChildren(parentCode)
  } catch (e) {
    children.value = []
    showApiError(e, '加载区划失败')
  } finally {
    listLoading.value = false
  }
}

function syncTreeCurrent(code: string) {
  void nextTick(() => {
    treeRef.value?.setCurrentKey(code || undefined)
  })
}

function syncUrl(code: string) {
  const current = typeof route.query.code === 'string' ? route.query.code : ''
  if (current === code) return
  void router.replace({ path: route.path, query: { ...route.query, code: code || undefined } })
}

async function openDetail(code: string) {
  if (!canViewDetail.value) {
    ElMessage.warning('暂无区划详情权限')
    return
  }
  detailVisible.value = true
  detailLoading.value = true
  detail.value = null
  try {
    const res = await getRegionDetail(code)
    detail.value = res.data
  } catch (e) {
    showApiError(e, '获取区划详情失败')
  } finally {
    detailLoading.value = false
  }
}

/** 从省级到该区划自身的完整路径（不依赖详情权限）。 */
async function fetchTrail(code: string): Promise<Crumb[]> {
  const res = await getRegionTrail(code)
  return (res.data.list || []).map(toCrumb)
}

async function setCurrentPath(trail: Crumb[], focusCode?: string) {
  stack.value = trail.length ? trail : [ROOT]
  const current = stack.value[stack.value.length - 1] ?? ROOT
  highlightCode.value = focusCode && focusCode !== current.code ? focusCode : ''
  detailVisible.value = false
  detail.value = null
  syncTreeCurrent(focusCode ?? current.code)
  syncUrl(current.code)
  await loadChildren(current.code)
}

async function expandTrail(codes: string[]) {
  await nextTick()
  for (const code of codes) {
    const node = treeRef.value?.getNode(code)
    if (node && !node.expanded && !node.isLeaf) {
      await new Promise<void>((resolve) => {
        node.expand(() => resolve())
      })
    }
  }
}

async function onTreeClick(data: TreeNode) {
  const crumb = toCrumb(data)
  const existIdx = stack.value.findIndex((c) => c.code === data.code)
  if (existIdx >= 0) {
    await setCurrentPath(stack.value.slice(0, existIdx + 1))
    return
  }
  const parentIdx = stack.value.findIndex((c) => c.code === data.parentCode)
  if (parentIdx >= 0) {
    await setCurrentPath([...stack.value.slice(0, parentIdx + 1), crumb])
    return
  }
  try {
    await setCurrentPath([ROOT, ...(await fetchTrail(data.code))])
  } catch (e) {
    showApiError(e, '定位区划失败')
  }
}

async function enterRegion(row: RegionItem) {
  const existIdx = stack.value.findIndex((c) => c.code === row.code)
  const trail = existIdx >= 0 ? stack.value.slice(0, existIdx + 1) : [...stack.value, toCrumb(row)]
  await setCurrentPath(trail)
  await expandTrail(trail.slice(1).map((c) => c.code))
}

async function enterFromDetail(row: RegionAdminItem) {
  if (row.level >= REGION_LEAF_LEVEL) return
  await enterRegion(row)
}

async function goCrumb(index: number) {
  await setCurrentPath(stack.value.slice(0, index + 1))
}

async function goRoot() {
  await setCurrentPath([ROOT])
}

/** 定位到命中项的上级列表并高亮命中行；有详情权限时同时打开详情。 */
async function locateFromSearch(row: RegionItem) {
  try {
    const trail = await fetchTrail(row.code)
    const ancestors = trail.slice(0, -1)
    await setCurrentPath([ROOT, ...ancestors], row.code)
    await expandTrail(ancestors.map((a) => a.code))
    syncTreeCurrent(row.code)
    if (canViewDetail.value) await openDetail(row.code)
  } catch (e) {
    showApiError(e, '定位区划失败')
  }
}

async function handleSearch() {
  const keyword = searchForm.keyword.trim()
  if (!keyword) {
    ElMessage.warning('请输入名称或区划码')
    return
  }
  searchLoading.value = true
  try {
    const res = await searchRegion({ keyword, level: searchForm.level, limit: SEARCH_LIMIT })
    searchHits.value = res.data.list || []
    searchTruncated.value = res.data.truncated
    if (!searchHits.value.length) ElMessage.info('未找到匹配区划')
  } catch (e) {
    showApiError(e, '搜索区划失败')
  } finally {
    searchLoading.value = false
  }
}

async function handleResetSearch() {
  searchForm.keyword = ''
  searchForm.level = undefined
  searchHits.value = []
  searchTruncated.value = false
  await goRoot()
}

onMounted(async () => {
  const code = typeof route.query.code === 'string' ? route.query.code.trim() : ''
  if (!code) {
    await loadChildren('')
    return
  }
  try {
    const trail = await fetchTrail(code)
    await setCurrentPath([ROOT, ...trail])
    await expandTrail(trail.map((c) => c.code))
  } catch (e) {
    showApiError(e, '定位区划失败')
    await goRoot()
  }
})
</script>

<style scoped>
.region-root-node {
  padding: 6px 8px;
  margin-bottom: 4px;
  border-radius: 4px;
  font-size: 14px;
  color: var(--el-text-color-primary);
  cursor: pointer;
  user-select: none;
}

.region-root-node:hover {
  background: var(--el-fill-color-light);
}

.region-root-node.is-active {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 600;
}
</style>
