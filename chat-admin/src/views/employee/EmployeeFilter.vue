<template>
  <el-form :inline="true" class="search-form" @submit.prevent>
    <el-form-item label="手机号">
      <el-input
        v-model="mobile"
        placeholder="请输入手机号"
        clearable
        style="width: 160px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="姓名">
      <el-input
        v-model="name"
        placeholder="请输入姓名"
        clearable
        style="width: 140px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="工号">
      <el-input
        v-model="employeeNo"
        placeholder="请输入工号"
        clearable
        style="width: 140px"
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
        placeholder="含下级部门"
        style="width: 180px"
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
      />
    </el-form-item>
    <el-form-item label="角色">
      <el-select v-model="roleId" placeholder="全部" clearable filterable style="width: 150px">
        <el-option v-for="r in roleOptions" :key="r.id" :label="r.name" :value="r.id" />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 110px">
        <el-option
          v-for="opt in STATUS_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="创建时间">
      <el-date-picker
        v-model="timeRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        :value-format="DATETIME_FORMAT"
        :shortcuts="DATE_RANGE_SHORTCUTS"
        :default-time="DATE_RANGE_DEFAULT_TIME"
        style="width: 360px"
      />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emit('search')">查询</el-button>
      <el-button @click="emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import {
  DATE_RANGE_DEFAULT_TIME,
  DATE_RANGE_SHORTCUTS,
  DATETIME_FORMAT,
  STATUS_OPTIONS,
} from '@/constants'
import type { DepartmentOption } from '@/api/department'
import type { EmployeeRoleBrief } from '@/api/employee'

defineProps<{
  departmentOptions: DepartmentOption[]
  roleOptions: EmployeeRoleBrief[]
}>()
const mobile = defineModel<string>('mobile', { required: true })
const name = defineModel<string>('name', { required: true })
const employeeNo = defineModel<string>('employeeNo', { required: true })
const departmentId = defineModel<number>('departmentId')
const roleId = defineModel<number>('roleId')
const status = defineModel<number>('status')
const timeRange = defineModel<[string, string] | null>('timeRange', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
