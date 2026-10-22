<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="帖子ID">
      <el-input
        v-model="idText"
        placeholder="帖子ID"
        clearable
        style="width: 110px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="标题/地点/内容"
        clearable
        style="width: 170px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="发布者">
      <el-input
        v-model="publisherKeyword"
        placeholder="昵称/手机号"
        clearable
        style="width: 140px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="发布者ID">
      <el-input
        v-model="publisherIdText"
        placeholder="用户ID"
        clearable
        style="width: 110px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="聊天区ID">
      <el-input
        v-model="roomIdText"
        placeholder="聊天区ID"
        clearable
        :disabled="standaloneOnly"
        style="width: 110px"
        @keyup.enter="emit('search')"
      />
      <el-checkbox v-model="standaloneOnly" class="inline-check">仅独立帖</el-checkbox>
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="category" placeholder="全部" clearable style="width: 100px">
        <el-option
          v-for="opt in POST_CATEGORY_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" placeholder="全部" clearable style="width: 100px">
        <el-option
          v-for="opt in POST_STATUS_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="发布时间">
      <el-date-picker
        v-model="timeRange"
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
    <el-form-item label="排序">
      <el-select v-model="sort" style="width: 110px">
        <el-option
          v-for="opt in POST_SORT_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item>
      <el-checkbox v-model="includeDeleted">含用户已删除</el-checkbox>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emit('search')">查询</el-button>
      <el-button @click="emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { computed, type Ref } from 'vue'

import { useDateRangeModel } from '@/composables/useDateRangeModel'
import { DATE_RANGE_DEFAULT_TIME, DATE_RANGE_SHORTCUTS, DATETIME_FORMAT } from '@/constants'
import { POST_CATEGORY_OPTIONS, POST_SORT_OPTIONS, POST_STATUS_OPTIONS } from '@/constants/post'

const id = defineModel<number>('id')
const keyword = defineModel<string>('keyword', { required: true })
const publisherId = defineModel<number>('publisherId')
const publisherKeyword = defineModel<string>('publisherKeyword', { required: true })
const roomId = defineModel<number>('roomId')
const category = defineModel<string>('category')
const status = defineModel<number>('status')
const startTime = defineModel<string>('startTime', { required: true })
const endTime = defineModel<string>('endTime', { required: true })
const sort = defineModel<'asc' | 'desc'>('sort', { required: true })
const includeDeleted = defineModel<boolean>('includeDeleted', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

/** 输入框里只接受正整数 ID，清空即不限。 */
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
const publisherIdText = positiveIdText(publisherId)
const roomIdText = positiveIdText(roomId)

const standaloneOnly = computed({
  get: () => roomId.value === 0,
  set: (v: boolean) => {
    roomId.value = v ? 0 : undefined
  },
})

const timeRange = useDateRangeModel(startTime, endTime)
</script>

<style scoped>
.inline-check {
  margin-left: 8px;
}
</style>
