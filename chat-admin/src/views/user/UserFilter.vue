<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="用户ID">
      <el-input
        v-model.number="id"
        placeholder="请输入用户ID"
        clearable
        style="width: 160px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="手机号">
      <el-input
        v-model="mobile"
        placeholder="请输入手机号"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="昵称">
      <el-input
        v-model="nickname"
        placeholder="请输入昵称"
        clearable
        style="width: 160px"
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
    <el-form-item label="禁言">
      <el-select v-model="muted" placeholder="全部" clearable style="width: 120px">
        <el-option label="已禁言" :value="true" />
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

const id = defineModel<number>('id')
const mobile = defineModel<string>('mobile', { required: true })
const nickname = defineModel<string>('nickname', { required: true })
const status = defineModel<number>('status')
const muted = defineModel<boolean>('muted')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
