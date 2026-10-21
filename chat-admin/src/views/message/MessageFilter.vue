<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="用户ID">
      <el-input
        v-model.number="userId"
        placeholder="发送者/接收者"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="对方ID">
      <el-input
        v-model.number="peerUserId"
        placeholder="私聊对方"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="聊天区ID">
      <el-input
        v-model.number="roomId"
        placeholder="房间ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="帖子ID">
      <el-input
        v-model.number="postId"
        placeholder="帖子ID"
        clearable
        style="width: 120px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="范围">
      <el-select v-model="scope" placeholder="全部" clearable style="width: 110px">
        <el-option
          v-for="opt in MESSAGE_SCOPE_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="消息内容"
        clearable
        style="width: 160px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="时间">
      <el-date-picker
        v-model="dateRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始"
        end-placeholder="结束"
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
import { MESSAGE_SCOPE_OPTIONS } from '@/constants/message'

const userId = defineModel<number>('userId')
const peerUserId = defineModel<number>('peerUserId')
const roomId = defineModel<number>('roomId')
const postId = defineModel<number>('postId')
const scope = defineModel<number>('scope')
const keyword = defineModel<string>('keyword', { required: true })
const dateRange = defineModel<[string, string] | null>('dateRange', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
