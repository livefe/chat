<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="ID">
      <el-input
        v-model="idText"
        placeholder="聊天区ID"
        clearable
        style="width: 110px"
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
      <el-select v-model="status" placeholder="全部" clearable style="width: 110px">
        <el-option
          v-for="item in ROOM_STATUS_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="来源">
      <el-select v-model="source" placeholder="全部" clearable style="width: 100px">
        <el-option
          v-for="item in ROOM_SOURCE_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="placeType" placeholder="全部" clearable style="width: 120px">
        <el-option
          v-for="item in ROOM_PLACE_TYPE_OPTIONS"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="创建人ID">
      <el-input
        v-model="userIdText"
        placeholder="用户ID"
        clearable
        style="width: 110px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emit('search')">查询</el-button>
      <el-button @click="emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { computed, type Ref } from 'vue'

import { ROOM_PLACE_TYPE_OPTIONS, ROOM_SOURCE_OPTIONS, ROOM_STATUS_OPTIONS } from '@/constants/room'

const id = defineModel<number>('id')
const name = defineModel<string>('name', { required: true })
const status = defineModel<number>('status')
const source = defineModel<number>('source')
const placeType = defineModel<number>('placeType')
const userId = defineModel<number>('userId')
const emit = defineEmits<{ search: []; reset: [] }>()

function positiveIdText(model: Ref<number | undefined>) {
  return computed({
    get: () => (model.value ? String(model.value) : ''),
    set: (v: string) => {
      const n = Number(v.trim())
      model.value = Number.isInteger(n) && n > 0 ? n : undefined
    },
  })
}

const idText = positiveIdText(id)
const userIdText = positiveIdText(userId)
</script>
