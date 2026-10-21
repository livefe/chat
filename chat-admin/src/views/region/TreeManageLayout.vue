<template>
  <div class="admin-page tree-manage-page">
    <el-card class="tree-manage-card">
      <template #header>
        <div class="card-header">
          <span>{{ title }}</span>
        </div>
      </template>

      <div v-if="$slots.search" class="tree-manage-search">
        <slot name="search" />
      </div>

      <slot name="hits" />

      <div class="tree-manage-body">
        <aside class="tree-manage-aside">
          <div class="tree-manage-aside__head">
            <span>{{ treeTitle }}</span>
          </div>
          <div class="tree-manage-aside__body">
            <slot name="tree" />
          </div>
        </aside>

        <section class="tree-manage-main">
          <div class="tree-manage-main__head">
            <div class="tree-manage-main__title">
              <slot name="breadcrumb" />
            </div>
            <div v-if="$slots['main-actions']" class="tree-manage-main__actions">
              <slot name="main-actions" />
            </div>
          </div>
          <div class="tree-manage-main__body" v-loading="mainLoading">
            <slot name="main" />
          </div>
        </section>
      </div>
    </el-card>

    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    treeTitle?: string
    mainLoading?: boolean
  }>(),
  {
    treeTitle: '结构树',
    mainLoading: false,
  },
)
</script>
