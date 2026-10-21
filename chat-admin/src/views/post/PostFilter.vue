<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="标题/地点"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="category" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in POST_CATEGORY_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="聊天区ID">
      <el-input
        v-model.number="roomId"
        placeholder="0=独立帖"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="发布者ID">
      <el-input
        v-model.number="publisherId"
        placeholder="用户ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in POST_STATUS_OPTIONS"
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
import { POST_CATEGORY_OPTIONS } from '@/constants/post'
import { POST_STATUS_OPTIONS } from '@/constants/post'

const keyword = defineModel<string>('keyword', { required: true })
const publisherId = defineModel<number>('publisherId')
const roomId = defineModel<number>('roomId')
const category = defineModel<string>('category')
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
