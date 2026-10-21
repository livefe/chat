<template>
  <el-form :inline="true" class="search-form">
    <el-form-item label="用户ID">
      <el-input
        v-model.number="userId"
        placeholder="参与用户"
        clearable
        style="width: 140px"
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
const userId = defineModel<number>('userId')
const emit = defineEmits<{ search: []; reset: [] }>()
</script>
