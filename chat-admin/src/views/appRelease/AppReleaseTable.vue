<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无版本记录">
    <el-table-column prop="platformText" label="平台" width="100" />
    <el-table-column prop="version" label="版本" width="120" />
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag v-if="row.isCurrent === 1" type="success" size="small">当前</el-tag>
        <el-tag v-else type="info" size="small">历史</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="大小" width="100">
      <template #default="{ row }">{{ formatSize(row.fileSize) }}</template>
    </el-table-column>
    <el-table-column prop="changelog" label="更新说明" min-width="160" show-overflow-tooltip />
    <el-table-column prop="downloadUrl" label="下载地址" min-width="180" show-overflow-tooltip />
    <el-table-column prop="createdAt" label="上传时间" width="170" />
    <el-table-column label="操作" width="220" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button size="small" link type="primary" @click="openDownload(row.downloadUrl)">
            下载
          </el-button>
          <el-button
            v-if="row.isCurrent !== 1 && can(PERM.OpsAppReleaseEnable)"
            size="small"
            link
            type="primary"
            @click="emit('enable', row as AppReleaseItem)"
          >
            设为当前
          </el-button>
          <el-button
            v-if="row.isCurrent === 1 && can(PERM.OpsAppReleaseDisable)"
            size="small"
            link
            type="danger"
            @click="emit('disable', row as AppReleaseItem)"
          >
            下架
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { AppReleaseItem } from '@/api/appRelease'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'

defineProps<{
  list: AppReleaseItem[]
  loading: boolean
}>()

const emit = defineEmits<{
  enable: [row: AppReleaseItem]
  disable: [row: AppReleaseItem]
}>()

const { can } = usePermission()

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function openDownload(url: string) {
  window.open(url, '_blank')
}
</script>
