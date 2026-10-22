<template>
  <el-form :inline="true" class="search-form" @submit.prevent>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="名称或编码"
        clearable
        style="width: 200px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="type" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="t in permissionTypeOptions"
          :key="t.value"
          :label="t.label"
          :value="t.value"
        />
      </el-select>
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
import { permissionTypeOptions } from '@/utils/permission'

const keyword = defineModel<string>('keyword', { required: true })
const type = defineModel<number>('type')
const status = defineModel<number>('status')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
