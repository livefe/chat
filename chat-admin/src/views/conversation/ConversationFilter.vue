<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="用户ID">
      <el-input
        v-model.number="userId"
        placeholder="参与用户"
        clearable
        style="width: 130px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="对方ID">
      <el-input
        v-model.number="peerUserId"
        placeholder="与用户ID同填查两人会话"
        clearable
        style="width: 190px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="最后活跃">
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

const userId = defineModel<number>('userId')
const peerUserId = defineModel<number>('peerUserId')
const startTime = defineModel<string>('startTime', { required: true })
const endTime = defineModel<string>('endTime', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const range = useDateRangeModel(startTime, endTime)
</script>
