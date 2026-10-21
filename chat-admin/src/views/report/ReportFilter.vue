<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="补充说明/处理说明"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="对象类型">
      <el-select v-model="targetType" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in REPORT_TARGET_TYPE_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="原因">
      <el-select v-model="reason" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in REPORT_REASON_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in REPORT_STATUS_OPTIONS"
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
import {
  REPORT_TARGET_TYPE_OPTIONS,
  REPORT_REASON_OPTIONS,
  REPORT_STATUS_OPTIONS,
} from '@/constants/report'
const keyword = defineModel<string>('keyword', { required: true })
const targetType = defineModel<number>('targetType')
const reason = defineModel<number>('reason')
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
