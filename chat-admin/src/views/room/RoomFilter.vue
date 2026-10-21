<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="ID">
      <el-input
        v-model.number="id"
        placeholder="聊天区ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="名称">
      <el-input
        v-model="name"
        placeholder="请输入名称"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 120px">
        <el-option label="待审核" :value="0" />
        <el-option label="正常" :value="1" />
        <el-option label="驳回" :value="2" />
        <el-option label="禁用" :value="3" />
      </el-select>
    </el-form-item>
    <el-form-item label="来源">
      <el-select v-model="source" placeholder="全部" clearable style="width: 120px">
        <el-option label="后台" :value="1" />
        <el-option label="用户" :value="2" />
      </el-select>
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="placeType" placeholder="全部" clearable style="width: 130px">
        <el-option
          v-for="item in ROOM_PLACE_TYPE_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
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
import { ROOM_PLACE_TYPE_OPTIONS } from '@/constants/room'

const id = defineModel<number>('id')
const name = defineModel<string>('name', { required: true })
const status = defineModel<number>('status')
const source = defineModel<number>('source')
const placeType = defineModel<number>('placeType')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
