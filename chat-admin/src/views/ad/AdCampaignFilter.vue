<template>
  <el-form :inline="true" class="search-form" @submit.prevent>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="广告主 / 标题 / 名称"
        clearable
        style="width: 180px"
        @keyup.enter="emit('search')"
        @clear="emit('search')"
      />
    </el-form-item>
    <el-form-item label="广告主">
      <el-select
        v-model="advertiserName"
        placeholder="全部"
        clearable
        filterable
        style="width: 160px"
        @change="emit('search')"
      >
        <el-option v-for="name in advertisers" :key="name" :label="name" :value="name" />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select
        v-model="status"
        placeholder="全部"
        clearable
        style="width: 120px"
        @change="emit('search')"
      >
        <el-option
          v-for="opt in AD_STATUS_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="排期">
      <el-date-picker
        v-model="scheduleRange"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始"
        end-placeholder="结束"
        unlink-panels
        :shortcuts="DATE_RANGE_SHORTCUTS"
        style="width: 240px"
        @change="emit('search')"
      />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="emit('search')">查询</el-button>
      <el-button @click="emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getAdAdvertisers } from '@/api/ad'
import { DATE_RANGE_SHORTCUTS } from '@/constants'
import { AD_STATUS_OPTIONS } from '@/constants/ad'

const keyword = defineModel<string>('keyword', { required: true })
const status = defineModel<number | undefined>('status')
const advertiserName = defineModel<string>('advertiserName', { required: true })
const startDate = defineModel<string>('startDate', { required: true })
const endDate = defineModel<string>('endDate', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const advertisers = ref<string[]>([])

const scheduleRange = computed({
  get: (): [string, string] | null =>
    startDate.value && endDate.value ? [startDate.value, endDate.value] : null,
  set: (v: [string, string] | null) => {
    startDate.value = v?.[0] ?? ''
    endDate.value = v?.[1] ?? ''
  },
})

onMounted(async () => {
  try {
    const { data } = await getAdAdvertisers()
    advertisers.value = data.list || []
  } catch {
    advertisers.value = []
  }
})
</script>
