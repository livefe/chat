<template>
  <el-drawer v-model="visible" title="区划详情" size="420px">
    <template v-if="detail">
      <el-descriptions :column="1" border v-loading="loading">
        <el-descriptions-item label="名称">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="简称">{{ detail.shortName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="层级">{{ detail.levelText }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'info'" size="small">
            {{ detail.statusText }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="区划码">{{ detail.code }}</el-descriptions-item>
        <el-descriptions-item label="上级区划码">{{
          detail.parentCode || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="全路径">{{ detail.fullPath }}</el-descriptions-item>
        <el-descriptions-item label="天地图国标码">{{
          detail.tiandituGb || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="高德 adcode">{{
          detail.amapAdcode || '-'
        }}</el-descriptions-item>
        <el-descriptions-item label="数据年份">{{ detail.dataYear || '-' }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="detail.level < 3" class="detail-actions">
        <el-button type="primary" @click="emit('enter', detail)">进入下级</el-button>
      </div>
    </template>
    <el-empty v-else-if="!loading" description="暂无详情权限或数据不存在" :image-size="72" />
  </el-drawer>
</template>

<script setup lang="ts">
import type { RegionAdminItem } from '@/api/region'

const visible = defineModel<boolean>({ required: true })
defineProps<{
  detail: RegionAdminItem | null
  loading: boolean
}>()
const emit = defineEmits<{ enter: [row: RegionAdminItem] }>()
</script>
