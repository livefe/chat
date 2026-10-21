<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="手机号">
      <el-input
        v-model="mobile"
        placeholder="请输入手机号"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="姓名">
      <el-input
        v-model="name"
        placeholder="请输入姓名"
        clearable
        style="width: 150px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="部门">
      <el-tree-select
        v-model="departmentId"
        :data="departmentOptions"
        check-strictly
        clearable
        filterable
        :render-after-expand="false"
        placeholder="请选择部门"
        style="width: 200px"
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
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
import type { DepartmentOption } from '@/api/department'

defineProps<{
  departmentOptions: DepartmentOption[]
}>()
const mobile = defineModel<string>('mobile', { required: true })
const name = defineModel<string>('name', { required: true })
const departmentId = defineModel<number>('departmentId')
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
