<script setup lang="ts">
import { computed, ref } from 'vue'

export type PostListAction = {
  label: string
  danger?: boolean
  disabled?: boolean
  onClick: () => void
}

const props = defineProps<{
  actions: PostListAction[]
}>()

const visible = ref(false)

const sheetActions = computed(() =>
  props.actions.map((action) => ({
    name: action.label,
    color: action.danger ? '#e05a52' : undefined,
    disabled: action.disabled,
  })),
)

function openSheet() {
  visible.value = true
}

function onSelect(item: { name: string }) {
  const action = props.actions.find((entry) => entry.label === item.name)
  action?.onClick()
}
</script>

<template>
  <button
    v-if="actions.length"
    type="button"
    class="post-action-menu__trigger"
    aria-label="管理"
    @click.stop="openSheet"
  >
    <van-icon name="ellipsis" size="18" />
  </button>
  <van-action-sheet
    v-model:show="visible"
    :actions="sheetActions"
    cancel-text="取消"
    close-on-click-action
    @select="onSelect"
  />
</template>

<style scoped>
.post-action-menu__trigger {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: var(--line-meta);
  margin: 0;
  padding: 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  -webkit-appearance: none;
  appearance: none;
}

.post-action-menu__trigger::before {
  content: '';
  position: absolute;
  inset: -10px -6px;
}
.post-action-menu__trigger:active {
  background: rgba(0, 0, 0, 0.05);
}
</style>
