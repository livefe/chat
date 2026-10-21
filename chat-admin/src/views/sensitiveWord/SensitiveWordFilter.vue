<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="词语">
      <el-input
        v-model="word"
        placeholder="搜索屏蔽词"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="分类">
      <el-select v-model="category" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in SENSITIVE_CATEGORY_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in STATUS_OPTIONS"
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
import { STATUS_OPTIONS } from '@/constants'
import { SENSITIVE_CATEGORY_OPTIONS } from '@/constants/sensitiveWord'

const word = defineModel<string>('word', { required: true })
const category = defineModel<string>('category', { required: true })
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
