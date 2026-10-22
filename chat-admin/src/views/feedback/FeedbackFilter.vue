<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="状态">
      <el-select v-model="status" style="width: 110px">
        <el-option
          v-for="opt in FEEDBACK_STATUS_FILTER_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="type" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="opt in FEEDBACK_TYPE_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="用户ID">
      <el-input
        v-model.number="userId"
        placeholder="提交用户"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="内容/联系方式/处理说明"
        clearable
        style="width: 200px"
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
import { FEEDBACK_STATUS_FILTER_OPTIONS, FEEDBACK_TYPE_OPTIONS } from '@/constants/feedback'

const status = defineModel<number>('status', { required: true })
const type = defineModel<number>('type')
const userId = defineModel<number>('userId')
const keyword = defineModel<string>('keyword', { required: true })
const startTime = defineModel<string>('startTime', { required: true })
const endTime = defineModel<string>('endTime', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const range = useDateRangeModel(startTime, endTime)
</script>
