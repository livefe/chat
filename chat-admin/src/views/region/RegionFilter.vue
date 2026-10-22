<template>
  <el-form :inline="true" class="search-form" @submit.prevent>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="名称/简称/区划码"
        clearable
        style="width: 200px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="层级">
      <el-select v-model="level" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in REGION_LEVEL_OPTIONS"
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
import { REGION_LEVEL_OPTIONS } from '@/constants/region'

const keyword = defineModel<string>('keyword', { required: true })
const level = defineModel<number>('level')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
