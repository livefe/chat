<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="广告主 / 标题"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in AD_STATUS_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emit('search')">查询</el-button>
      <el-button @click="emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { AD_STATUS_OPTIONS } from '@/constants/ad'
const keyword = defineModel<string>('keyword', { required: true })
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
