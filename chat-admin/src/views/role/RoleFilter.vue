<template>
  <el-form :inline="true" class="search-form" @submit.prevent>
    <el-form-item label="角色名称">
      <el-input
        v-model="name"
        placeholder="请输入角色名称"
        clearable
        style="width: 170px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="角色编码">
      <el-input
        v-model="code"
        placeholder="请输入角色编码"
        clearable
        style="width: 170px"
        @keyup.enter="emit('search')"
      />
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

const name = defineModel<string>('name', { required: true })
const code = defineModel<string>('code', { required: true })
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
