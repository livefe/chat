<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="操作人ID">
      <el-input
        v-model.number="employeeId"
        placeholder="员工ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="模块">
      <el-select v-model="module" placeholder="全部" clearable style="width: 140px">
        <el-option
          v-for="opt in moduleOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="操作">
      <el-select v-model="action" placeholder="全部" clearable style="width: 140px">
        <el-option
          v-for="opt in actionOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="操作人/目标/详情"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="时间范围">
      <el-date-picker
        v-model="timeRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        value-format="YYYY-MM-DD HH:mm:ss"
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
import type { AuditActionOption, AuditModuleOption } from '@/api/auditLog'

defineProps<{
  moduleOptions: AuditModuleOption[]
  actionOptions: AuditActionOption[]
}>()

const employeeId = defineModel<number>('employeeId')
const module = defineModel<string>('module', { required: true })
const action = defineModel<string>('action', { required: true })
const keyword = defineModel<string>('keyword', { required: true })
const timeRange = defineModel<[string, string] | null>('timeRange', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
