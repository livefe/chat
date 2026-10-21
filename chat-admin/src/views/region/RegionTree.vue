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
        v-if="canViewDetail && currentNode.code && currentNode.level < 3"
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
import { useRoute } from 'vue-router'
import { ElMessage, type LoadFunction, type TreeInstance } from 'element-plus'
import {
  getRegionChildren,
  getRegionDetail,
  searchRegion,
  type RegionAdminItem,
  type RegionItem,
} from '@/api/region'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'
import RegionChildTable from './RegionChildTable.vue'
import RegionDetailDrawer from './RegionDetailDrawer.vue'
import RegionFilter from './RegionFilter.vue'
import RegionSearchHits from './RegionSearchHits.vue'
import TreeManageLayout from './TreeManageLayout.vue'

type Crumb = { code: string; name: string; level: number }
type TreeNode = RegionItem & { isLeaf?: boolean }

const ROOT: Crumb = { code: '', name: '全国', level: 0 }

const { can } = usePermission()
const route = useRoute()
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

const stack = ref<Crumb[]>([ROOT])
const children = ref<RegionItem[]>([])
const listLoading = ref(false)
const detail = ref<RegionAdminItem | null>(null)
const detailLoading = ref(false)
const detailVisible = ref(false)

const currentNode = computed(() => stack.value[stack.value.length - 1] ?? ROOT)

function toTreeNodes(list: RegionItem[]): TreeNode[] {
  return list.map((item) => ({
    ...item,
    isLeaf: item.level >= 3,
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
  nextTick(() => {
    treeRef.value?.setCurrentKey(code || undefined)
  })
}

async function openDetail(code: string) {
  if (!canViewDetail.value) {
    ElMessage.warning('暂无区划详情权限')
    return
  }
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await getRegionDetail(code)
    detail.value = res.data
  } catch (e) {
    detail.value = null
    showApiError(e, '获取区划详情失败')
  } finally {
    detailLoading.value = false
  }
}

async function buildAncestorTrail(item: RegionItem): Promise<Crumb[]> {
  const trail: Crumb[] = []
  let code = item.parentCode
  while (code) {
    const res = await getRegionDetail(code)
    trail.unshift({
      code: res.data.code,
      name: res.data.name,
      level: res.data.level,
    })
    code = res.data.parentCode
  }
  return trail
}

async function setCurrentPath(trail: Crumb[], focusCode?: string) {
  stack.value = trail.length ? trail : [ROOT]
  const current = stack.value[stack.value.length - 1] ?? ROOT
  detailVisible.value = false
  detail.value = null
  syncTreeCurrent(focusCode ?? current.code)
  await loadChildren(current.code)
}

async function onTreeClick(data: TreeNode) {
  const crumb: Crumb = { code: data.code, name: data.name, level: data.level }
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
  if (!data.parentCode) {
    await setCurrentPath([ROOT, crumb])
    return
  }
  try {
    const ancestors = await buildAncestorTrail(data)
    await setCurrentPath([ROOT, ...ancestors, crumb])
  } catch {
    await setCurrentPath([ROOT, crumb])
  }
}

async function enterRegion(row: RegionItem) {
  const crumb: Crumb = { code: row.code, name: row.name, level: row.level }
  const existIdx = stack.value.findIndex((c) => c.code === row.code)
  const trail = existIdx >= 0 ? stack.value.slice(0, existIdx + 1) : [...stack.value, crumb]
  await setCurrentPath(trail)
  nextTick(() => {
    const node = treeRef.value?.getNode(row.code)
    if (node && !node.expanded) node.expand()
  })
}

async function enterFromDetail(row: RegionAdminItem) {
  if (row.level >= 3) return
  await enterRegion(row)
}

async function goCrumb(index: number) {
  await setCurrentPath(stack.value.slice(0, index + 1))
}

async function goRoot() {
  await setCurrentPath([ROOT])
}

async function expandTrail(codes: string[]) {
  await nextTick()
  for (const code of codes) {
    const node = treeRef.value?.getNode(code)
    if (node && !node.expanded) {
      await new Promise<void>((resolve) => {
        node.expand(() => resolve())
      })
    }
  }
}

async function locateFromSearch(row: RegionItem) {
  try {
    if (canViewDetail.value) {
      const ancestors = await buildAncestorTrail(row)
      await setCurrentPath([ROOT, ...ancestors])
      await expandTrail(ancestors.map((a) => a.code))
      await openDetail(row.code)
    } else if (row.parentCode) {
      const parentCrumb: Crumb = {
        code: row.parentCode,
        name: row.fullPath?.split('/').slice(0, -1).join('/') || row.parentCode,
        level: Math.max(0, row.level - 1),
      }
      await setCurrentPath([ROOT, parentCrumb])
      ElMessage.info('已定位到上级，当前账号无详情权限')
    } else {
      await setCurrentPath([ROOT])
      ElMessage.info('已定位，当前账号无详情权限')
    }
  } catch (e) {
    showApiError(e, '定位区划失败')
  }
}

async function handleSearch() {
  if (!searchForm.keyword.trim()) {
    ElMessage.warning('请输入搜索关键词')
    return
  }
  searchLoading.value = true
  try {
    const res = await searchRegion({
      keyword: searchForm.keyword.trim(),
      level: searchForm.level,
      limit: 50,
    })
    searchHits.value = res.data.list || []
    if (!searchHits.value.length) {
      ElMessage.info('未找到匹配区划')
    }
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
  await goRoot()
}

onMounted(() => loadChildren(''))
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
