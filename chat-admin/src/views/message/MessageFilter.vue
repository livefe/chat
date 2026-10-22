<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="消息ID">
      <el-input
        v-model.number="id"
        placeholder="消息ID"
        clearable
        style="width: 110px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="用户ID">
      <el-input
        v-model.number="userId"
        placeholder="发送或接收"
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
      >
        <template #suffix>
          <el-tooltip
            content="与用户ID同时填写时只看两人之间的私聊；单独填写时查该用户的全部私聊"
            placement="top"
          >
            <el-icon class="hint-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </template>
      </el-input>
    </el-form-item>
    <el-form-item label="聊天区ID">
      <el-input
        v-model.number="roomId"
        placeholder="聊天区ID"
        clearable
        style="width: 110px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="帖子ID">
      <el-input
        v-model.number="postId"
        placeholder="帖子ID"
        clearable
        style="width: 110px"
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
  <div v-if="keywordOnly" class="filter-hint">
    未指定用户、聊天区、帖子或时间时，关键词只搜索最近 7 天的消息。
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { QuestionFilled } from '@element-plus/icons-vue'

import { useDateRangeModel } from '@/composables/useDateRangeModel'
import { DATE_RANGE_DEFAULT_TIME, DATE_RANGE_SHORTCUTS, DATETIME_FORMAT } from '@/constants'
import { MESSAGE_SCOPE_OPTIONS } from '@/constants/message'

const id = defineModel<number>('id')
const userId = defineModel<number>('userId')
const peerUserId = defineModel<number>('peerUserId')
const roomId = defineModel<number>('roomId')
const postId = defineModel<number>('postId')
const scope = defineModel<number>('scope')
const keyword = defineModel<string>('keyword', { required: true })
const startTime = defineModel<string>('startTime', { required: true })
const endTime = defineModel<string>('endTime', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const range = useDateRangeModel(startTime, endTime)
const keywordOnly = computed(
  () =>
    !!keyword.value.trim() &&
    !id.value &&
    !userId.value &&
    !peerUserId.value &&
    !roomId.value &&
    !postId.value &&
    !startTime.value,
)
</script>

<style scoped>
.hint-icon {
  cursor: help;
}

.filter-hint {
  margin: -6px 0 12px;
  font-size: 12px;
  color: var(--el-color-warning);
}
</style>
