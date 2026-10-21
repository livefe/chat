<script setup lang="ts">
defineProps<{
  tabs?: { key: string; label: string; count?: number; warn?: boolean }[]
  activeTab?: string
}>()

const emit = defineEmits<{
  'update:activeTab': [key: string]
}>()
</script>

<template>
  <div class="im-shell">
    <aside class="im-sidebar">
      <div v-if="tabs?.length" class="im-sidebar__tabs">
        <button
          v-for="t in tabs"
          :key="t.key"
          type="button"
          class="im-tab"
          :class="{ 'is-active': activeTab === t.key }"
          @click="emit('update:activeTab', t.key)"
        >
          <span>{{ t.label }}</span>
          <span
            v-if="t.count != null && t.count > 0"
            class="im-tab__count"
            :class="{ 'is-warn': t.warn }"
          >
            {{ t.count > 99 ? '99+' : t.count }}
          </span>
        </button>
      </div>
      <div v-if="$slots.search" class="im-sidebar__search">
        <slot name="search" />
      </div>
      <div class="im-sidebar__list">
        <slot name="list" />
      </div>
    </aside>
    <section class="im-main">
      <slot name="main" />
    </section>
  </div>
</template>

<style src="./im-sidebar.css"></style>
<style src="./im-thread.css"></style>
<style src="./im-composer.css"></style>
<style src="./im-responsive.css"></style>
