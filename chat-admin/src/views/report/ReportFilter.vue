<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="状态">
      <el-select v-model="status" style="width: 110px">
        <el-option
          v-for="opt in REPORT_STATUS_FILTER_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="对象类型">
      <el-select v-model="targetType" placeholder="全部" clearable style="width: 110px">
        <el-option
          v-for="opt in REPORT_TARGET_TYPE_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="对象ID">
      <el-input
        v-model.number="targetId"
        placeholder="消息/帖子/用户ID"
        clearable
        style="width: 150px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="被举报用户">
      <el-input
        v-model.number="targetUserId"
        placeholder="用户ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="举报人">
      <el-input
        v-model.number="reporterId"
        placeholder="用户ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="原因">
      <el-select v-model="reason" placeholder="全部" clearable style="width: 110px">
        <el-option
          v-for="opt in REPORT_REASON_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="补充说明/处理说明"
        clearable
        style="width: 170px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="提交时间">
      <el-date-picker
        v-model="range"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始"
        end-placeholder="结束"
        :value-format="DATETIME_FORMAT"
        :shortcuts="DATE_RANGE_SHORTCUTS"
        :default-time="DATE_RANGE_DEFAULT_TIME"
        style="width: 340px"
      />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emit('search')">查询</el-button>
      <el-button @click="emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { useDateRangeModel } from '@/composables/useDateRangeModel'
import { DATE_RANGE_DEFAULT_TIME, DATE_RANGE_SHORTCUTS, DATETIME_FORMAT } from '@/constants'
import {
  REPORT_REASON_OPTIONS,
  REPORT_STATUS_FILTER_OPTIONS,
  REPORT_TARGET_TYPE_OPTIONS,
} from '@/constants/report'

const status = defineModel<number>('status', { required: true })
const targetType = defineModel<number>('targetType')
const targetId = defineModel<number>('targetId')
const targetUserId = defineModel<number>('targetUserId')
const reporterId = defineModel<number>('reporterId')
const reason = defineModel<number>('reason')
const keyword = defineModel<string>('keyword', { required: true })
const startTime = defineModel<string>('startTime', { required: true })
const endTime = defineModel<string>('endTime', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const range = useDateRangeModel(startTime, endTime)
</script>
