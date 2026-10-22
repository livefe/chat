<template>
  <el-form :inline="true" class="search-form" @submit.prevent>
    <el-form-item label="操作人">
      <el-select
        v-model="employeeId"
        filterable
        remote
        clearable
        :remote-method="searchEmployees"
        :loading="employeeLoading"
        placeholder="姓名/手机号/工号"
        style="width: 190px"
      >
        <el-option
          v-for="e in employeeOptions"
          :key="e.id"
          :label="e.name || e.mobile"
          :value="e.id"
        >
          <span>{{ e.name || '-' }}</span>
          <span class="option-sub">{{ e.mobile }}</span>
        </el-option>
      </el-select>
    </el-form-item>
    <el-form-item label="模块">
      <el-select v-model="module" placeholder="全部" clearable filterable style="width: 140px">
        <el-option
          v-for="opt in options.modules"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
    </el-form-item>
    <el-form-item label="操作">
      <el-select v-model="action" placeholder="全部" clearable filterable style="width: 170px">
        <el-option
          v-for="opt in actionOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        >
          <span>{{ opt.label }}</span>
          <span v-if="!module" class="option-sub">{{ moduleLabel(opt.module) }}</span>
        </el-option>
      </el-select>
    </el-form-item>
    <el-form-item label="结果">
      <el-select v-model="success" placeholder="全部" clearable style="width: 100px">
        <el-option label="成功" :value="true" />
        <el-option label="失败" :value="false" />
      </el-select>
    </el-form-item>
    <el-form-item label="对象">
      <el-select v-model="targetType" placeholder="类型" clearable filterable style="width: 120px">
        <el-option
          v-for="opt in options.targetTypes"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-input-number
        v-model="targetId"
        :min="1"
        :controls="false"
        placeholder="对象 ID"
        class="target-id-input"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="接口">
      <el-input
        v-model="requestPath"
        placeholder="请求路径片段"
        clearable
        style="width: 200px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="关键词">
      <el-input
        v-model="keyword"
        placeholder="操作人/对象/路径/详情"
        clearable
        style="width: 190px"
        @keyup.enter="emit('search')"
      />
    </el-form-item>
    <el-form-item label="时间">
      <el-date-picker
        v-model="timeRange"
        type="datetimerange"
        range-separator="至"
        start-placeholder="开始时间"
        end-placeholder="结束时间"
        :value-format="DATETIME_FORMAT"
        :shortcuts="DATE_RANGE_SHORTCUTS"
        :default-time="DATE_RANGE_DEFAULT_TIME"
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
import { computed, ref, watch } from 'vue'
import { searchAuditEmployees, type AuditEmployeeOption, type AuditOptions } from '@/api/auditLog'
import { DATE_RANGE_DEFAULT_TIME, DATE_RANGE_SHORTCUTS, DATETIME_FORMAT } from '@/constants'

const props = defineProps<{ options: AuditOptions }>()

const employeeId = defineModel<number>('employeeId')
const module = defineModel<string>('module', { required: true })
const action = defineModel<string>('action', { required: true })
const targetType = defineModel<string>('targetType', { required: true })
const targetId = defineModel<number>('targetId')
const requestPath = defineModel<string>('requestPath', { required: true })
const success = defineModel<boolean>('success')
const keyword = defineModel<string>('keyword', { required: true })
const timeRange = defineModel<[string, string] | null>('timeRange', { required: true })
const emit = defineEmits<{ search: []; reset: [] }>()

const actionOptions = computed(() =>
  module.value
    ? props.options.actions.filter((a) => a.module === module.value)
    : props.options.actions,
)

function moduleLabel(value: string) {
  return props.options.modules.find((m) => m.value === value)?.label ?? ''
}

watch(module, () => {
  if (action.value && !actionOptions.value.some((a) => a.value === action.value)) {
    action.value = ''
  }
})

const employeeOptions = ref<AuditEmployeeOption[]>([])
const employeeLoading = ref(false)
let searchSeq = 0

async function searchEmployees(kw: string) {
  const seq = ++searchSeq
  employeeLoading.value = true
  try {
    const { data } = await searchAuditEmployees({ keyword: kw })
    if (seq === searchSeq) employeeOptions.value = data
  } catch {
    if (seq === searchSeq) employeeOptions.value = []
  } finally {
    if (seq === searchSeq) employeeLoading.value = false
  }
}

watch(
  employeeId,
  async (id) => {
    if (!id || employeeOptions.value.some((e) => e.id === id)) return
    try {
      const { data } = await searchAuditEmployees({ id })
      employeeOptions.value = [...data, ...employeeOptions.value]
    } catch {
      // 回显失败时下拉框仅显示 ID，不影响筛选
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.option-sub {
  float: right;
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.target-id-input {
  width: 110px;
  margin-left: 6px;
}
</style>
