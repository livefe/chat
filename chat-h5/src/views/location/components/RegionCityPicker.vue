<script setup lang="ts">
import { ref, watch } from 'vue'

import { getRegionChildren, searchRegion, type RegionItem } from '@/api/region'
import { showCatchError } from '@/utils/toast'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  select: [region: RegionItem]
}>()

const keyword = ref('')
const loading = ref(false)
const list = ref<RegionItem[]>([])
const stack = ref<RegionItem[]>([])

let searchTimer: ReturnType<typeof setTimeout> | null = null

function close() {
  emit('update:show', false)
}

function displayName(item: RegionItem) {
  return item.name
}

async function loadChildren(parentCode = '') {
  loading.value = true
  try {
    const res = await getRegionChildren(parentCode)
    list.value = res.data.list
  } catch (e: unknown) {
    list.value = []
    showCatchError(e, '加载区划失败', true)
  } finally {
    loading.value = false
  }
}

async function runSearch(value: string) {
  const text = value.trim()
  if (!text) {
    await loadChildren(stack.value.at(-1)?.code ?? '')
    return
  }
  loading.value = true
  try {
    const res = await searchRegion({ keyword: text, limit: 30 })
    list.value = res.data.list
  } catch (e: unknown) {
    list.value = []
    showCatchError(e, '搜索失败', true)
  } finally {
    loading.value = false
  }
}

function onKeywordInput(value: string) {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    void runSearch(value)
  }, 300)
}

async function openItem(item: RegionItem) {
  if (item.level === 1) {
    loading.value = true
    try {
      const res = await getRegionChildren(item.code)
      if (!res.data.list.length) {
        emit('select', item)
        close()
        return
      }
      stack.value = [...stack.value, item]
      keyword.value = ''
      list.value = res.data.list
    } catch (e: unknown) {
      showCatchError(e, '加载区划失败', true)
    } finally {
      loading.value = false
    }
    return
  }
  emit('select', item)
  close()
}

function goBack() {
  if (!stack.value.length) return
  stack.value = stack.value.slice(0, -1)
  keyword.value = ''
  void loadChildren(stack.value.at(-1)?.code ?? '')
}

watch(
  () => props.show,
  (visible) => {
    if (!visible) return
    keyword.value = ''
    stack.value = []
    void loadChildren('')
  },
)
</script>

<template>
  <van-popup
    :show="show"
    position="bottom"
    round
    :style="{ height: '72vh' }"
    @update:show="emit('update:show', $event)"
  >
    <div class="city-picker">
      <div class="app-sheet__handle city-picker__handle" />
      <div class="city-picker__header">
        <button v-if="stack.length" type="button" class="city-picker__back" @click="goBack">
          <van-icon name="arrow-left" />
          返回
        </button>
        <span class="city-picker__title">选择城市</span>
        <button type="button" class="city-picker__close" @click="close">关闭</button>
      </div>

      <van-search
        v-model="keyword"
        placeholder="搜索省/市/区县"
        shape="round"
        @update:model-value="onKeywordInput"
        @search="runSearch(keyword)"
      />

      <div v-if="stack.length" class="city-picker__crumb">
        <span v-for="(item, index) in stack" :key="item.code" class="city-picker__crumb-item">
          {{ displayName(item) }}<span v-if="index < stack.length - 1"> / </span>
        </span>
      </div>

      <div class="city-picker__list">
        <van-loading v-if="loading" class="city-picker__loading">加载中…</van-loading>
        <van-empty v-else-if="!list.length" description="暂无区划" />
        <button
          v-for="item in list"
          :key="item.code"
          type="button"
          class="city-picker__item"
          @click="openItem(item)"
        >
          <div class="city-picker__item-main">
            <span class="city-picker__item-name">{{ displayName(item) }}</span>
          </div>
          <van-icon v-if="item.level === 1" name="arrow" class="city-picker__item-arrow" />
        </button>
      </div>
    </div>
  </van-popup>
</template>

<style scoped>
.city-picker {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--card-bg);
  padding-top: var(--sheet-pad-top);
}

.city-picker__handle {
  margin-bottom: 4px;
}

.city-picker__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 16px 0;
}

.city-picker__title {
  font-size: var(--sheet-title-size);
  font-weight: var(--sheet-title-weight);
  color: var(--text-primary);
}

.city-picker__back,
.city-picker__close {
  border: none;
  background: transparent;
  color: var(--brand-color-dark);
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.city-picker__list {
  flex: 1;
  overflow-y: auto;
}

.city-picker__loading {
  display: flex;
  justify-content: center;
  padding: 24px;
}

.city-picker__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
  padding: var(--sheet-item-padding-y) 16px;
  border: none;
  border-bottom: 1px solid var(--border-color);
  background: transparent;
  text-align: left;
}

.city-picker__item:active {
  background: rgba(0, 0, 0, 0.04);
}

.city-picker__item-main {
  min-width: 0;
}

.city-picker__item-name {
  display: block;
  font-size: var(--sheet-item-size);
  color: var(--text-primary);
  line-height: 1.5;
  font-weight: var(--sheet-item-weight);
}

.city-picker__item-arrow {
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.city-picker__crumb {
  padding: 8px 16px 10px;
  font-size: 12px;
  color: #888888;
}

.city-picker__crumb-item:last-child {
  color: #2a6bd8;
  font-weight: 600;
}
</style>
