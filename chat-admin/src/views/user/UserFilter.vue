<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="用户ID">
      <el-input
        v-model.number="id"
        placeholder="请输入用户ID"
        clearable
        style="width: 140px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="手机号">
      <el-input
        v-model="mobile"
        placeholder="完整或部分号码，含已注销"
        clearable
        style="width: 200px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="昵称">
      <el-input
        v-model="nickname"
        placeholder="请输入昵称"
        clearable
        style="width: 150px"
        @keyup.enter="emit('search')"
      />
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
    <el-form-item label="禁言">
      <el-select v-model="muted" placeholder="全部" clearable style="width: 110px">
        <el-option label="禁言中" :value="true" />
      </el-select>
    </el-form-item>
    <el-form-item label="注册时间">
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
import {
  DATE_RANGE_DEFAULT_TIME,
  DATE_RANGE_SHORTCUTS,
  DATETIME_FORMAT,
  STATUS_OPTIONS,
} from '@/constants'

const id = defineModel<number>('id')
const mobile = defineModel<string>('mobile', { required: true })
const nickname = defineModel<string>('nickname', { required: true })
const status = defineModel<number>('status')
const muted = defineModel<boolean>('muted')
const startTime = defineModel<string>('startTime', { required: true })
const endTime = defineModel<string>('endTime', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const range = useDateRangeModel(startTime, endTime)
</script>
